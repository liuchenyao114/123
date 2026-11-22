import DOMPurify from 'dompurify';
import { marked } from 'marked';

// 所有 dangerouslySetInnerHTML 的内容必须先经过这里净化，
// 否则消息内容里的 <img onerror> 等可执行脚本会窃取 localStorage 中的令牌。
export function renderSafeMarkdown(content) {
  if (!content) return '';
  const html = marked.parse(String(content));
  return DOMPurify.sanitize(html);
}
