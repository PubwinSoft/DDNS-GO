const LANG = localStorage.getItem('lang') || (navigator.language || navigator.browserLanguage).replaceAll('_', '-').toLowerCase();

// 添加新的语言字典
const addI18n = (arg1, arg2) => {
  if (typeof arg1 === 'string') {
    I18N_MAP[arg1] = arg2;
  } else {
    for (let key in arg1) {
      I18N_MAP[key] = arg1[key];
    }
  }
}

// 支持两种调用方式：
// 1. 文本的key + (可选：语言映射字典)，{en: {hello: "hello", world: "world"}, zh: {hello: "你好", world: "世界"}}
// 2. 语言字符串字典，{en: "hello", zh: "你好"}
const i18n = (key, langMap = I18N_MAP) => {
  if (typeof key !== 'string') {
    langMap = key;
    key = null;
  }
  // 优先取地区语言，否则取表示语言，再否则取表示语言相同的地区语言，最后取英文
  let lang = 'en';
  if (LANG in langMap) {
    lang = LANG;
  } else if (LANG.split('-')[0] in langMap) {
    lang = LANG.split('-')[0];
  } else {
    for (const l in langMap) {
      if (l.split('-')[0] === LANG.split('-')[0]) {
        lang = l;
        break;
      }
    }
  }
  let text = '';
  if (key) {
    text = langMap[lang][key];
  } else {
    text = langMap[lang];
  }
  if (text === undefined) {
    console.warn(`i18n: No translation for ${key}`);
    return key;
  }
  return text;
}

const convertDom = (dom = document, ...args) => {
  $('[data-i18n]', dom).each((_,el) => {
    const key = $(el).data('i18n');
    $(el).text(i18n(key, ...args));
  });
  $('[data-i18n_html]', dom).each((_,el) => {
    const key = $(el).data('i18n_html');
    $(el).html(i18n(key, ...args));
  });
}

$(() => {convertDom();});