package src

import (
	"fmt"
	"io"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/hamba/avro"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Loader struct {
	table      string
	schema     avro.Schema
	encoder    *avro.Encoder
	pipewriter io.WriteCloser
	pipereader io.Reader

	errs chan error
}

func NewLoader(db sqlx.Execer, table string, schema avro.Schema) *Loader {
	pr, pw := io.Pipe()
	encoder := avro.NewEncoderForSchema(schema, pw)
	errs := make(chan error)

	go func() {
		defer close(errs)
		err := LoadAvroStream(db, table, schema, pr)
		if err != nil {
			pr.CloseWithError(err)
		}
		errs <- err
	}()

	return &Loader{
		table:      table,
		schema:     schema,
		encoder:    encoder,
		pipewriter: pw,
		pipereader: pr,
		errs:       errs,
	}
}

func (l *Loader) Encode(row interface{}) error {
	return l.encoder.Encode(row)
}

func (l *Loader) Close() error {
	// close the writer
	err := l.pipewriter.Close()
	if err != nil {
		return err
	}

	// wait for the loader to finish
	return <-l.errs
}

func LoadAvroStream(db sqlx.Execer, table string, schema avro.Schema, stream io.Reader) error {
	if schema.Type() != avro.Record {
		log.Fatal("only records can be loaded")
	}
	rec := schema.(*avro.RecordSchema)

	var columnMap []string
	for _, field := range rec.Fields() {
		columnMap = append(columnMap, fmt.Sprintf("%s <- %s", field.Name(), field.Name()))
	}
	sort.Strings(columnMap)

	readID := uuid.NewV4().String()
	query := fmt.Sprintf(`
		LOAD DATA LOCAL INFILE 'Reader::%s'
		REPLACE INTO TABLE %s
		FORMAT AVRO
		( %s )
		SCHEMA ?
		ERRORS HANDLE ?
	`, readID, table, strings.Join(columnMap, ", "))

	mysql.RegisterReaderHandler(readID, func() io.Reader { return stream })
	defer mysql.DeregisterReaderHandler(readID)

	_, err := db.Exec(query, schema.String(), table)
	return err
}

func AvroSchemaFromStruct(m interface{}) (avro.Schema, error) {
	mType := reflect.TypeOf(m)

	if mType.Kind() == reflect.Ptr {
		mType = mType.Elem()
	}

	if mType.Kind() != reflect.Struct {
		return nil, errors.New("can only generate Avro schema for a struct")
	}

	fields := make([]*avro.Field, 0, mType.NumField())
	for i := 0; i < mType.NumField(); i++ {
		f := mType.Field(i)
		fType := f.Type

		var schemaType avro.Type
		var nullable bool

		if fType.Kind() == reflect.Ptr {
			fType = fType.Elem()
			nullable = true
		}

		switch fType.Kind() {
		case reflect.String:
			schemaType = avro.String
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			schemaType = avro.Int
		case reflect.Int64:
			schemaType = avro.Long
		case reflect.Float32:
			schemaType = avro.Float
		case reflect.Float64:
			schemaType = avro.Double
		case reflect.Bool:
			schemaType = avro.Boolean
		default:
			log.Fatalf("type not supported: %s", fType.Kind())
		}

		var fieldSchema avro.Schema = avro.NewPrimitiveSchema(schemaType, nil)
		var err error
		if nullable {
			fieldSchema, err = avro.NewUnionSchema([]avro.Schema{fieldSchema, &avro.NullSchema{}})
			if err != nil {
				return nil, err
			}
		}

		field, err := avro.NewField(f.Name, fieldSchema, avro.NoDefault)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}

	schema, err := avro.NewRecordSchema(mType.Name(), "com.singlestore", fields)
	if err != nil {
		return nil, err
	}

	return schema, err
}
