import React from 'react';
import { Switch, Route, Link } from 'react-router-dom';

import Home from './Home';
import Replay from './Replay';

import './App.css';

const App: React.FC = () => {
    return (
        <div className="App">
            <div className="header">
                <Link to="/">Home</Link>
            </div>
            <Switch>
                <Route exact path="/" component={Home} />
                <Route path="/replay/:gameid" component={Replay} />
            </Switch>
        </div>
    );
};

export default App;
