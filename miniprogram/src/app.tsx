import React, { useEffect } from 'react';
import { useDidShow, useDidHide } from '@tarojs/taro';
import { useAppStore } from './store/app';
import './app.scss';

function App(props) {
  const init = useAppStore(state => state.init);

  useEffect(() => {
    init();
  }, []);

  useDidShow(() => {
    init();
  });

  useDidHide(() => {});

  return props.children;
}

export default App;
