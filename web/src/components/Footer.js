import React, { useContext, useEffect, useState } from 'react';

import { Container, Segment } from 'semantic-ui-react';
import { StatusContext } from '../context/Status';

const Footer = () => {
  const [Footer, setFooter] = useState('');
  const [statusState] = useContext(StatusContext);
  useEffect(() => {
    let savedFooter = localStorage.getItem('footer_html');
    if (!savedFooter) savedFooter = '';
    setFooter(savedFooter);
  });

  // 版本号以服务端 /api/status 为准，不再使用构建时注入的 REACT_APP_VERSION。
  // 否则前端产物与后端二进制版本不一致时，页脚会一直显示旧版本。
  let version =
    statusState && statusState.status && statusState.status.version;
  if (!version) {
    try {
      const savedStatus = JSON.parse(localStorage.getItem('status') || '{}');
      version = savedStatus.version;
    } catch (e) {
      version = '';
    }
  }
  if (!version) version = process.env.REACT_APP_VERSION || '';

  return (
    <Segment vertical>
      <Container textAlign='center'>
        {Footer === '' ? (
          <div className='custom-footer'>
            <a
              href='https://github.com/liuchenyao114/123'
              target='_blank'
            >
              消息推送服务 {version}{' '}
            </a>
            由{' '}
            <a href='https://github.com/songquanpeng' target='_blank'>
              JustSong
            </a>{' '}
            构建，源代码遵循{' '}
            <a href='https://opensource.org/licenses/mit-license.php'>
              MIT 协议
            </a>
          </div>
        ) : (
          <div
            className='custom-footer'
            dangerouslySetInnerHTML={{ __html: Footer }}
          ></div>
        )}
      </Container>
    </Segment>
  );
};

export default Footer;
