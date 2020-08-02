package main

import (
	"bytes"
	"encoding/base32"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/cretz/bine/torutil/ed25519"
)

type v3Address struct {
	PublicKey, PrivateKey []byte
	OnionAddress          string
}

var done = make(chan struct{})

func cancel() bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}

var strExp string

var match struct {
	sync.Mutex
	count int
}

func init() {
	flag.StringVar(&strExp, "r", "", "regex to match")
	flag.IntVar(&match.count, "c", 1, "number of matches")
}

func main() {
	flag.Parse()

	address := make(chan *v3Address)
	re, _ := regexp.Compile(strExp)
	var n sync.WaitGroup
	cpu := runtime.NumCPU()

	if match.count < cpu && len(strExp) < match.count {
		cpu = match.count
	}

	for x := 0; x < cpu; x++ {
		n.Add(1)
		go genAddress(&n, address)
	}

	go func() {
		n.Wait()
		close(address)
	}()

	for {
		select {
		case <-done:
			for range address {
			}
			return
		case addr, _ := <-address:
			a := addr
			go addrMatch(a, re)
		}
	}

}

func genAddress(n *sync.WaitGroup, out chan<- *v3Address) {
	defer n.Done()
	for {
		if cancel() {
			return
		}
		addr := new(v3Address)
		ed, _ := ed25519.GenerateKey(nil)
		addr.PublicKey, addr.PrivateKey = ed.PublicKey(), ed.PrivateKey()
		addr.OnionAddress = base32.StdEncoding.EncodeToString(addr.PublicKey)
		addr.OnionAddress = strings.ToLower(addr.OnionAddress[:52])
		out <- addr
	}
}

func addrMatch(a *v3Address, re *regexp.Regexp) {
	if match.count == 0 {
		return
	}
	if re.MatchString(a.OnionAddress) == true {
		match.Lock()
		defer match.Unlock()
		writeAddress(a)
		match.count -= 1
		if match.count == 0 {
			close(done)
		}
	}
}

func writeAddress(a *v3Address) {
	path := "v3/" + a.OnionAddress
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	f1, _ := os.Create(path + "/hs_ed25519_secret_key")
	defer f1.Close()
	f2, _ := os.Create(path + "/hs_ed25519_public_key")
	defer f2.Close()
	d := fmtHeader("== ed25519v1-secret: type0 ==", a.PrivateKey)
	f1.Write(d)
	d = fmtHeader("== ed25519v1-public: type0 ==", a.PublicKey)
	f2.Write(d)
}

func fmtHeader(f string, data []byte) []byte {
	var header bytes.Buffer
	header.WriteString(f)
	header.Write([]byte{0x00, 0x00, 0x00})
	header.Write(data)
	return header.Bytes()
}
