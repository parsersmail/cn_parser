package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type CNParser struct {
	IP   string
	Body string
}

var inputFileName = flag.String("i", "", "имя входящего файла")
var cn = flag.String("cn", "", "CN not contains cn")
var verbose = flag.Bool("v", false, "print log")

func fileNameWithoutExtTrimSuffix(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func containsKeys(word string, keys []string) bool {
	for _, k := range keys {
		if strings.Contains(word, k) {
			return true
		}
	}
	return false
}

// go build -o cn_parser386.exe
//go build -o bin/cn_parser
//./csvCNparser -i a.csv -cn "api.nikatv.ru, example"
////GOOS=windows GOARCH=386 go build -o bin/cn_parser386.exe main.go
func main() {

	flag.Parse()

	var keys []string

	if *cn != "" {
		keys = strings.Split(*cn, ",")
		for i := 0; i < len(keys); {
			if keys[i] == "" {
				keys = append(keys[:i], keys[i+1:]...)
				continue
			}
			i++
		}
	}

	if *verbose {
		fmt.Println("///keys start show///")
		for _, v := range keys {
			fmt.Println(v)
		}
		fmt.Println("/////keys end show///")
	}

	var uniqueCN = make(map[string]struct{}, 0)
	var uniqueWithoutCN = make(map[string]struct{}, 0)

	f, err := os.Open(*inputFileName)
	if err != nil {
		log.Fatal(err)
	}

	var separator rune
	separator = '\t'
	csvReader := csv.NewReader(f)
	csvReader.Comma = separator
	csvReader.Comma = separator

	var ipAddr *CNParser
	var ipAddrs []CNParser

	for line, err := csvReader.Read(); err == nil; line, err = csvReader.Read() {
		addr := net.ParseIP(line[0])
		if addr != nil {
			if ipAddr != nil && strings.Contains(ipAddr.Body, "CN") {
				ipAddrs = append(ipAddrs, *ipAddr)
			}
			ipAddr = nil
			ipAddr = &CNParser{
				IP:   addr.String(),
				Body: "",
			}
		} else {
			if ipAddr != nil {
				ipAddr.Body += line[0] + "\n"
			}
		}
	}

	if ipAddr != nil && strings.Contains(ipAddr.Body, "CN") {
		ipAddrs = append(ipAddrs, *ipAddr)
	}

	fn := fileNameWithoutExtTrimSuffix(*inputFileName)
	outPostfix := fn + "_out.csv"
	outCnPostfix := fn + "_cn_out.csv"
	uniqueOutPostfix := fn + "_unique_out.csv"
	uniqueCnPostfix := fn + "_unique_cn_out.csv"

	fOut, err := os.OpenFile(outPostfix, os.O_CREATE|os.O_APPEND|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}

	defer fOut.Close()

	fCnOut, err := os.OpenFile(outCnPostfix, os.O_CREATE|os.O_APPEND|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer fCnOut.Close()

	UniqueOut, err := os.OpenFile(uniqueOutPostfix, os.O_CREATE|os.O_APPEND|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer UniqueOut.Close()

	UniqueCNOut, err := os.OpenFile(uniqueCnPostfix, os.O_CREATE|os.O_APPEND|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer UniqueCNOut.Close()

	for _, v := range ipAddrs {
		body := v.Body

		for _, bodyVal := range strings.Split(body, "\n") {
			if idxS := strings.Index(bodyVal, "subject"); idxS != -1 {

				if idx := strings.Index(bodyVal, "CN="); idxS != -1 {

					idxStart := idx + len("CN=")
					idxEnd := strings.Index(bodyVal[idxStart:], ",")
					if idxEnd == -1 {
						idxEnd = len(bodyVal)
					} else {
						idxEnd = idxStart + strings.Index(bodyVal[idxStart:], ",")
					}

					var n int
					var err error

					//filter with CN
					if len(keys) > 0 && containsKeys(bodyVal[idxStart:idxEnd], keys) {

						if strings.Contains(bodyVal[idxStart:idxEnd], " ") {
							n, err = fCnOut.WriteString(v.IP + "\t" + "\"" + bodyVal[idxStart:idxEnd] + "\"" + "\r\n")
							uniqueCN[bodyVal[idxStart:idxEnd]] = struct{}{}
						} else {
							n, err = fCnOut.WriteString(v.IP + "\t" + bodyVal[idxStart:idxEnd] + "\r\n")
							uniqueCN[bodyVal[idxStart:idxEnd]] = struct{}{}
						}

						if *verbose {
							log.Println("write to fCnOut ", n, "bytes: err =", err, "for string:", bodyVal[idxStart:idxEnd])
						}
						break
					}

					if strings.Contains(bodyVal[idxStart:idxEnd], " ") {
						n, err = fOut.WriteString(v.IP + "\t" + "\"" + bodyVal[idxStart:idxEnd] + "\"" + "\r\n")
					} else {
						n, err = fOut.WriteString(v.IP + "\t" + bodyVal[idxStart:idxEnd] + "\r\n")
					}

					uniqueWithoutCN[bodyVal[idxStart:idxEnd]] = struct{}{}

					if *verbose {
						log.Println("write to fOut", n, "bytes: err =", err, "for string:", bodyVal[idxStart:idxEnd])
					}
					break
				}
			}
		}
	}

	var keysCN []string

	for k, _ := range uniqueCN {
		keysCN = append(keysCN, k)
	}

	sort.Strings(keysCN)

	for _, k := range keysCN {

		if strings.Contains(k, " ") {
			UniqueCNOut.WriteString("\"" + k + "\"" + "\r\n")
		} else {
			UniqueCNOut.WriteString(k + "\r\n")
		}

	}

	var keysOut []string

	for k, _ := range uniqueWithoutCN {
		keysOut = append(keysOut, k)
	}

	sort.Strings(keysOut)

	for _, k := range keysOut {

		if strings.Contains(k, " ") {
			UniqueOut.WriteString("\"" + k + "\"" + "\r\n")
		} else {
			UniqueOut.WriteString(k + "\r\n")
		}

	}

	if *verbose {
		fmt.Println("uniqueWithoutCN", uniqueWithoutCN)
		fmt.Println("uniqueCN", uniqueCN)
	}

}
