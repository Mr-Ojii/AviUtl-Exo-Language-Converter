# AviUtl-Exo-Language-Converter
AviUtl(拡張編集)のオブジェクトファイルであるexoの言語変換用ソフトウェア

## 使い方
conf.tomlに
```toml
src = "変換元の言語"
dst = "変換先の言語"
```
を記載し、実行ファイルに変換したいexoファイルをD&Dします。

## lang.tomlについて
言語変換情報記述ファイルです。
- descriptionは概要について (もともとは改行コードや文字コードの設定も追加する気であった)
- description.languageは言語のリストを記述します
- mapsは言語変換情報の配列です
- maps[].nameにはexoの_nameのvalueに対応する文字列をdescription.languageのindexと対応するよう配列で記述します
- maps[].keysにはname下におけるkeyの変換情報をdescription.languageのindexと対応するよう配列で記述します

その場でパパっと修正できるよう後から読み込む形式にしましたが、誤りを見つけた場合はIssueやプルリクをいただけると嬉しいです。
