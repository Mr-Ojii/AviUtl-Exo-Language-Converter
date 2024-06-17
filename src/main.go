package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type config struct {
	Src string `toml:"src"`
	Dst string `toml:"dst"`
}

type description struct {
	Language []string `toml:"Language"`
}

type maps struct {
	Name []string   `toml:"name"`
	Keys [][]string `toml:"keys"`
}

type langmap struct {
	Desc description `toml:"description"`
	Maps []maps      `toml:"maps"`
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func getLangIndex(conf *config, lang *langmap) (int, int) {
	var srci, dsti = -1, -1
	for i, l := range lang.Desc.Language {
		if conf.Src == l {
			srci = i
		}
		if conf.Dst == l {
			dsti = i
		}
	}
	if srci == -1 || dsti == -1 {
		panic("Failed to get index of language")
	}

	return srci, dsti
}

func sjisToUtf8(sjis []byte) []byte {
	buf := bytes.NewBuffer(sjis)
	rio := transform.NewReader(buf, japanese.ShiftJIS.NewDecoder())
	ret, err := io.ReadAll(rio)
	if err != nil {
		panic(err)
	}
	return ret
}

func utf8ToSjis(utf8 string) []byte {
	buf := strings.NewReader(utf8)
	rio := transform.NewReader(buf, japanese.ShiftJIS.NewEncoder())
	ret, err := io.ReadAll(rio)
	if err != nil {
		panic(err)
	}
	return ret
}

func main() {
	if len(os.Args) < 2 {
		panic("too few argments")
	}
	inputFilename := os.Args[1]
	if !fileExists(inputFilename) {
		panic("file not exists")
	}

	// confを読むため、Chdir
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	os.Chdir(filepath.Dir(exePath))

	var conf config
	_, err = toml.DecodeFile("conf.toml", &conf)
	if err != nil {
		panic("failed to decode config.toml")
	}

	var lang langmap
	_, err = toml.DecodeFile("lang.toml", &lang)
	if err != nil {
		panic("failed to decode lang.toml")
	}

	srci, dsti := getLangIndex(&conf, &lang)

	// .exoで終わるのであれば、4文字-したスライスでよい
	outputFilename := inputFilename[:len(inputFilename)-4] + "_" + conf.Dst + ".exo"
	outfile, err := os.Create(outputFilename)
	if err != nil {
		panic("failed to open output file")
	}
	defer outfile.Close()
	fmt.Println("Export to " + outputFilename)

	origBytes, err := os.ReadFile(inputFilename)
	if err != nil {
		panic("failed to read input file")
	}
	origString := string(sjisToUtf8(origBytes))

	// CRLF/CRをLFにしておく
	origString = strings.ReplaceAll(strings.ReplaceAll(origString, "\r\n", "\n"), "\r", "\n")

	// LFでSplit
	splitedString := strings.Split(origString, "\n")

	convString := ""

	var Maps maps

	// ぐちゃぐちゃ...
	for _, ss := range splitedString {
		if len(ss) != 0 {
			// セクション名であれば、そのまま出力してMapをリセット
			if ss[0] == '[' {
				convString += ss
				Maps.Name = nil
				Maps.Keys = nil
			} else {
				//それ以外なら = で cut
				bef, aft, found := strings.Cut(ss, "=")
				// みつからなかった場合、空行や誰かが書いたコメントである可能性があるため、そのまま
				if !found {
					convString += ss
				} else {
					// _nameだったら探して、Mapを登録
					if bef == "_name" {
						for _, mp := range lang.Maps {
							if mp.Name[srci] == aft {
								Maps = mp
								aft = mp.Name[dsti]
								break
							}
						}
					}

					// KeyをMapから探して、マッチするものがあれば挿げ替えて変換
					for _, mp := range Maps.Keys {
						if bef == mp[srci] {
							bef = mp[dsti]
							break
						}
					}

					convString += bef + "=" + aft
				}
			}
		}
		convString += "\r\n"
	}

	outfile.Write(utf8ToSjis(convString))
}
