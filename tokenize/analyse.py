import sys
from io import StringIO

from fugashi import Tagger


def main():
  tagger = Tagger('-Owakati')
  # 先尝试从命令行提取文本，没有就依靠用户输入
  try:
    text = sys.argv[1]
  except IndexError:
    text = input()
  # text = "冴えない彼女の育てかた"
  tagger.parse(text)
  f = StringIO()
  for word in tagger(text):
    print(word, word.feature.lemma, word.feature.pronBase, word.pos.replace(',*', ''), sep='\t', file=f)
  print(f.getvalue())
  f.close()


if __name__ == '__main__':
  main()
