import React from 'react';
import ReactDOM from 'react-dom';
import { BrowserRouter as Router } from 'react-router-dom';
import { Switch, Route, Link } from 'react-router-dom';

import Home from './Home';
import Replay from './Replay';

import 'tailwindcss/tailwind.css';

ReactDOM.render(
    <React.StrictMode>
        <Router>
            <div className="relative bg-white">
                <div className="px-4">
                    <Switch>
                        <Route exact path="/" component={Home} />
                        <Route path="/replay/:gameid" component={Replay} />
                    </Switch>
                </div>
            </div>
        </Router>
    </React.StrictMode>,
    document.getElementById('root')
);
