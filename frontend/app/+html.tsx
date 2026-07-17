import { ScrollViewStyleReset } from 'expo-router/html';
import { type PropsWithChildren } from 'react';

export default function Root({ children }: PropsWithChildren) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta httpEquiv="X-UA-Compatible" content="IE=edge" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0, user-scalable=no" />
        <meta name="theme-color" content="#000000" />
        <link rel="manifest" href="/manifest.json" />
        <ScrollViewStyleReset />
        <style dangerouslySetInnerHTML={{ __html: `
          * {
            touch-action: pan-x pan-y !important;
            -ms-touch-action: pan-x pan-y !important;
            -webkit-user-select: none !important;
            -moz-user-select: none !important;
            -ms-user-select: none !important;
            user-select: none !important;
          }
          
          html, body {
            overflow: hidden !important;
            height: 100% !important;
            width: 100% !important;
            position: fixed !important;
            top: 0 !important;
            left: 0 !important;
          }
          
          #root {
            height: 100% !important;
            width: 100% !important;
            overflow: hidden !important;
          }
        ` }} />
        <script dangerouslySetInnerHTML={{ __html: `
          document.addEventListener('gesturestart', function (e) {
            e.preventDefault();
          }, { passive: false });
          
          document.addEventListener('gesturechange', function (e) {
            e.preventDefault();
          }, { passive: false });
          
          document.addEventListener('gestureend', function (e) {
            e.preventDefault();
          }, { passive: false });
          
          document.addEventListener('touchstart', function (e) {
            if (e.touches.length > 1) {
              e.preventDefault();
            }
          }, { passive: false });
          
          let lastTouchEnd = 0;
          document.addEventListener('touchend', function (e) {
            const now = Date.now();
            if (now - lastTouchEnd <= 300) {
              e.preventDefault();
            }
            lastTouchEnd = now;
          }, { passive: false });
          
          document.addEventListener('wheel', function (e) {
            if (e.ctrlKey || e.metaKey) {
              e.preventDefault();
            }
          }, { passive: false });
          
          document.addEventListener('keydown', function (e) {
            if ((e.ctrlKey || e.metaKey) && (e.key === '+' || e.key === '-' || e.key === '=' || e.key === '0')) {
              e.preventDefault();
            }
          });
        ` }} />
      </head>
      <body>{children}</body>
    </html>
  );
}
