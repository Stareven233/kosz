from fugashi import Tagger


tagger = Tagger('-Owakati')
# text = "麩菓子は、麩を主材料とした日本の菓子。"
text = "冴えない彼女の育てかた"
tagger.parse(text)
# => '麩 菓子 は 、 麩 を 主材 料 と し た 日本 の 菓子 。'
for word in tagger(text):
  print(word, word.feature.lemma, word.feature.pronBase, word.pos.replace(',*', ''), sep='\t')
  # "feature" is the Unidic feature data as a named tuple

# 没文档也没源文件，凑合着用
dir_word = [
  'char_type', 'feature', 'feature_raw', 'is_unk', 'length', 'pos', 'posid', 'rlength', 'stat', 'surface',
  'white_space'
]
dir_word_feature = [
  '_asdict', '_field_defaults', '_fields', '_make', '_replace', 'aConType', 'aModType', 'aType', 'cForm', 'cType',
  'count', 'fConType', 'fForm', 'fType', 'form', 'formBase', 'goshu', 'iConType', 'iForm', 'iType', 'index', 'kana',
  'kanaBase', 'lForm', 'lemma', 'lemma_id', 'lid', 'orth', 'orthBase', 'pos1', 'pos2', 'pos3', 'pos4', 'pron',
  'pronBase', 'type'
]
