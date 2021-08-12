import React from 'react';
import ReactDOM from 'react-dom';
import { BrowserRouter as Router } from 'react-router-dom';
import { Switch, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from 'react-query';

import Home from './Home';
import Replay from './Replay';

import 'tailwindcss/tailwind.css';

const queryClient = new QueryClient();

ReactDOM.render(
    <React.StrictMode>
        <Router>
            <QueryClientProvider client={queryClient}>
                <div className="relative bg-white">
                    <div className="px-4">
                        <Switch>
                            <Route exact path="/" component={Home} />
                            <Route path="/replay/:gameid" component={Replay} />
                        </Switch>
                    </div>
                </div>
            </QueryClientProvider>
        </Router>
    </React.StrictMode>,
    document.getElementById('root')
);
