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

func file_exists(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil
}

func get_lang_index(conf *config, lang *langmap) (int, int) {
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

func sjis_to_utf8(sjis []byte) []byte {
    buf := bytes.NewBuffer(sjis)
    rio := transform.NewReader(buf, japanese.ShiftJIS.NewDecoder())
    ret, err := io.ReadAll(rio)
    if err != nil {
        panic(err)
    }
    return ret
}

func utf8_to_sjis(utf8 string) []byte {
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
        panic("Too few argments")
    }
    input_filename := os.Args[1]
    if !file_exists(input_filename) {
        panic("File not exists")
    }

    // confを読むため、Chdir
    exe_path, err := os.Executable()
    if err != nil {
        panic(err)
    }
	os.Chdir(filepath.Dir(exe_path))

    var conf config
    _, err = toml.DecodeFile("conf.toml", &conf)
    if err != nil {
        panic("Failed to decode config.toml")
    }

    var lang langmap
    _, err = toml.DecodeFile("lang.toml", &lang)
    if err != nil {
        panic("Failed to decode lang.toml")
    }

    srci, dsti := get_lang_index(&conf, &lang)

    // .exoで終わるのであれば、4文字-したスライスでよい
    output_filename := input_filename[:len(input_filename)-4] + "_" + conf.Dst + ".exo"
    outfile, err := os.Create(output_filename)
    if err != nil {
        panic("Failed to open output file")
    }
    defer outfile.Close()
    fmt.Println("Export to " + output_filename)

    orig_bytes, err := os.ReadFile(input_filename)
    if err != nil {
        panic("Failed to read input file")
    }
    orig_string := string(sjis_to_utf8(orig_bytes))

    // CRLF/CRをLFにしておく
    orig_string = strings.ReplaceAll(strings.ReplaceAll(orig_string, "\r\n", "\n"), "\r", "\n")

    // LFでSplit
    splited_string := strings.Split(orig_string, "\n")

    conv_string := ""

    var Maps maps

    // ぐちゃぐちゃ...
    for _, ss := range splited_string {
        if len(ss) != 0 {
            // セクション名であれば、そのまま出力してMapをリセット
            if ss[0] == '[' {
                conv_string += ss
                Maps.Name = nil
                Maps.Keys = nil
            } else {
                //それ以外なら = で cut
                bef, aft, found := strings.Cut(ss, "=")
                // みつからなかった場合、空行や誰かが書いたコメントである可能性があるため、そのまま
                if !found {
                    conv_string += ss
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

                    conv_string += bef + "=" + aft
                }
            }
        }
        conv_string += "\r\n"
    }

    outfile.Write(utf8_to_sjis(conv_string))
}
