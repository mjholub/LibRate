const calcEntropy = (charset: string, length: number) => {
  Math.round(Math.log2(Math.pow(charset.length, length)));
}

const stdCharsets = [{
  name: 'lowercase',
  re: /[a-z]/,
  length: 26,
}, {
  name: 'uppercase',
  re: /[A-Z]/,
  length: 26,
}, {
  name: 'numbers',
  re: /[0-9]/,
  length: 10,
}, {
  name: 'symbols',
  re: /[^a-zA-Z0-9]/,
  length: 33,
}]

const calcCharsetLength = (charset: string) => {
  const charsetLength = stdCharsets.reduce((acc, charset) => {
    return acc + (charset.re.test(charset) ? charset.length : 0);
  }, 0);
  return charsetLength;
};

