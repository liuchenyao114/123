import React, { lazy, Suspense, useContext, useEffect } from 'react';
import { Route, Routes } from 'react-router-dom';
import Loading from './components/Loading';
import User from './pages/User';
import { PrivateRoute } from './components/PrivateRoute';
import RegisterForm from './components/RegisterForm';
import LoginForm from './components/LoginForm';
import NotFound from './pages/NotFound';
import Setting from './pages/Setting';
import EditUser from './pages/User/EditUser';
import AddUser from './pages/User/AddUser';
import { API, showError } from './helpers';
import PasswordResetForm from './components/PasswordResetForm';
import GitHubOAuth from './components/GitHubOAuth';
import PasswordResetConfirm from './components/PasswordResetConfirm';
import { UserContext } from './context/User';
import { StatusContext } from './context/Status';
import Message from './pages/Message';
import Channel from './pages/Channel';
import EditChannel from './pages/Channel/EditChannel';
import EditMessage from './pages/Message/EditMessage';
import Webhook from './pages/Webhook';
import EditWebhook from './pages/Webhook/EditWebhook';

const Home = lazy(() => import('./pages/Home'));
const About = lazy(() => import('./pages/About'));

function App() {
  const [userState, userDispatch] = useContext(UserContext);
  const [statusState, statusDispatch] = useContext(StatusContext);

  const loadUser = () => {
    let user = localStorage.getItem('user');
    if (user) {
      let data = JSON.parse(user);
      userDispatch({ type: 'login', payload: data });
    }
  };
  const loadStatus = async () => {
    const res = await API.get('/api/status');
    const { success, data } = res.data;
    if (success) {
      localStorage.setItem('status', JSON.stringify(data));
      statusDispatch({ type: 'set', payload: data });
      localStorage.setItem('footer_html', data.footer_html);
      localStorage.setItem('home_page_link', data.home_page_link);
      // 后端版本是唯一真源。比较「前端构建版本」与服务端版本：
      // 若前端落后，通常是因为 CDN 或浏览器缓存了旧的 index.html，导致一直加载
      // 旧版脚本。此时仅提示「Shift+F5」无效（CDN 按 URL 缓存，强刷仍命中旧缓存），
      // 因此自动换一个带版本号的 URL 重载：新 URL 是全新的缓存 key，会强制回源
      // 拿到最新首页。URL 已带该版本号时不再跳转，避免死循环。
      const serverVersion = (data.version || '').replace(/^v/, '');
      const buildVersion = (process.env.REACT_APP_VERSION || '').replace(/^v/, '');
      if (serverVersion && serverVersion !== '0.0.0' && buildVersion) {
        if (buildVersion !== serverVersion) {
          const url = new URL(window.location.href);
          if (url.searchParams.get('_v') !== serverVersion) {
            url.searchParams.set('_v', serverVersion);
            window.location.replace(url.toString());
            return;
          }
        }
        localStorage.setItem('app_version', serverVersion);
      }
    } else {
      showError('无法正常连接至服务器！');
    }
  };

  useEffect(() => {
    loadUser();
    loadStatus().then();
  }, []);

  return (
    <Routes>
      <Route
        path='/'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <Home />
          </Suspense>
        }
      />
      <Route
        path='/user'
        element={
          <PrivateRoute>
            <User />
          </PrivateRoute>
        }
      />
      <Route
        path='/user/edit/:id'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <EditUser />
          </Suspense>
        }
      />
      <Route
        path='/user/edit'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <EditUser />
          </Suspense>
        }
      />
      <Route
        path='/user/add'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <AddUser />
          </Suspense>
        }
      />
      <Route
        path='/user/reset'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <PasswordResetConfirm />
          </Suspense>
        }
      />
      <Route
        path='/channel'
        element={
          <PrivateRoute>
            <Channel />
          </PrivateRoute>
        }
      />
      <Route
        path='/channel/edit/:id'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <EditChannel />
          </Suspense>
        }
      />
      <Route
        path='/channel/add'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <EditChannel />
          </Suspense>
        }
      />
      <Route
        path='/webhook'
        element={
          <PrivateRoute>
            <Webhook />
          </PrivateRoute>
        }
      />
      <Route
        path='/webhook/edit/:id'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <EditWebhook />
          </Suspense>
        }
      />
      <Route
        path='/webhook/add'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <EditWebhook />
          </Suspense>
        }
      />
      <Route
        path='/message'
        element={
          <PrivateRoute>
            <Message />
          </PrivateRoute>
        }
      />
      <Route
        path='/editor'
        element={
          <PrivateRoute>
            <EditMessage />
          </PrivateRoute>
        }
      />
      <Route
        path='/editor/:id'
        element={
          <PrivateRoute>
            <EditMessage />
          </PrivateRoute>
        }
      />
      <Route
        path='/login'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <LoginForm />
          </Suspense>
        }
      />
      <Route
        path='/register'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <RegisterForm />
          </Suspense>
        }
      />
      <Route
        path='/reset'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <PasswordResetForm />
          </Suspense>
        }
      />
      <Route
        path='/oauth/github'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <GitHubOAuth />
          </Suspense>
        }
      />
      <Route
        path='/setting'
        element={
          <PrivateRoute>
            <Suspense fallback={<Loading></Loading>}>
              <Setting />
            </Suspense>
          </PrivateRoute>
        }
      />
      <Route
        path='/about'
        element={
          <Suspense fallback={<Loading></Loading>}>
            <About />
          </Suspense>
        }
      />
      <Route path='*' element={<NotFound />} />
    </Routes>
  );
}

export default App;
