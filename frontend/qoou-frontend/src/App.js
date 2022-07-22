import React from 'react';
import { Button } from 'antd';
import './App.css';
import { Login } from './components/login';

const App = () => (
  <div className="App">
    <Button type="primary">Button</Button>
    <Login />
  </div>
);

export default App;