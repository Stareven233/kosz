$Env:CGO_LDFLAGS = "-L/E:/MeCab/sdk -L/E:/MeCab/bin/libmecab.dll -L/E:/MeCab/mecab-main/mecab/src -lmecab -lstdc++"
$Env:CGO_CFLAGS = " -I/E:/MeCab/sdk -I/E:/MeCab/mecab-main/mecab/src"
go build -o test.exe main.go
# 垃圾sm东西，改了半天各种报错层出不穷，根本用不了
