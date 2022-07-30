from janome.tokenizer import Tokenizer

t = Tokenizer()
for token in t.tokenize('あなた何を見惚れているのかしら'):
  print(token)

