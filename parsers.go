/*
   Copyright 2017 Mike Lloyd

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const (
	headerLength = 512

	// I'll be blatantly honest, I have no idea what these numbers are actually for?
	// Now that I"ve thought about it...I think they are the large number that a signature has to reach in order to
	// be validated? Or they are one part of the equation used to validate a signature? I should really figure that
	// part out.
	nStr      = "118640995551645342603070001658453189751527774412027743746599405743243142607464144767361060640655844749760788890022283424922762488917565551002467771109669598189410434699034532232228621591089508178591428456220796841621637175567590476666928698770143328137383952820383197532047771780196576957695822641224262693037"
	eStr      = "100001027"
	refString = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz+/"
)

var refCharMap = []string{
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l",
	"m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x",
	"y", "z",
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L",
	"M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X",
	"Y", "Z",
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"+", "/",
}

// ParseCVD reads the ClamAV CVD file, parses it to a struct in-memory, and then validates it. It returns a map of errors,
// if there are any. The error map contains [field]error.
func ParseCVD(b []byte, e *[]error) *ClamAV {
	var header []byte
	var def []byte
	header = append(header, b[0:headerLength]...)
	def = append(def, b[headerLength:]...)

	head := NewHeaders(header, def)

	if len(head.Problems) > 0 {
		*e = append(*e, head.Problems...)
		return &ClamAV{}
	}

	return &ClamAV{
		Header: head,
		Definition: AVDefinition{
			Body: def,
		},
	}
}

// NewHeaders parses the heder of a ClamAV definition file.
func NewHeaders(h, b []byte) HeaderFields {
	return parseHeader(h, b)
}

func newEmptyHeader() *HeaderFields {
	return &HeaderFields{
		Problems: make([]error, 0),
	}
}

func parseHeader(h, b []byte) HeaderFields {
	var errs []error
	hFields := HeaderFields{
		Problems: errs,
	}

	headStr := string(h)
	headParts := strings.Split(headStr, ":")
	if len(headParts) < 3 {
		hFields.Problems = append(hFields.Problems, errors.New("bad def header"))
	}

	hFields.ParseTime(headParts[1])
	hFields.Version = hFields.Atou(headParts[2])
	hFields.Signatures = hFields.Atou(headParts[3])
	hFields.Functionality = hFields.Atou(headParts[4])
	hFields.ParseMD5(headParts[5], b)
	hFields.Builder = headParts[7]

	return hFields
}

// ParseTime parses the build time signature.
func (h *HeaderFields) ParseTime(s string) {
	// Mon Jan 2 15:04:05 -0700 MST 2006
	pTime, err := time.Parse("02 Jan 2006 15-04 -0700", s)
	if err != nil {
		h.Problems = append(h.Problems, err)
	}
	h.CreationTime = pTime
}

// Atou is a simple helper function.
func (h *HeaderFields) Atou(s string) uint {
	x, err := strconv.Atoi(s)
	if err != nil {
		h.Problems = append(h.Problems, err)
	}
	return uint(x)
}

// ParseMD5 reads and validates the MD5 checksum of the AV definition.
func (h *HeaderFields) ParseMD5(md string, b []byte) {
	localHash := fmt.Sprintf("%x", md5.Sum(b))

	// this is the simple way around golang not having isalnum.
	// https://git.io/vyrcB
	if md != localHash {
		h.Problems = append(h.Problems, errors.New("md5 does not match"))
		h.MD5Valid = false
		h.MD5Hash = localHash
	}

	h.MD5Hash = md
	h.MD5Valid = true
}

func (h *HeaderFields) parseDSig(b []byte) {
	n, e := big.NewInt(0), big.NewInt(0)

	if err := readRadix(n, nStr, 10); err != nil {
		h.Problems = append(h.Problems, err)
	}

	if err := readRadix(e, eStr, 10); err != nil {
		h.Problems = append(h.Problems, err)
	}

	plainSig := h.decodeSig(string(b), 16, e, n)

	log.Debugf("plain sig: %s", plainSig)

	// libclamav only compares the first 16 characters. /shrug
	hexSig := hex.EncodeToString([]byte(string(plainSig[:16])))
	log.Debugf("decoded signature: %s", hexSig)

	if h.MD5Hash == hexSig {
		h.DSignature = hexSig
		h.DSigValid = true
	} else {
		h.DSignature = hexSig
		h.DSigValid = false
	}
}

// decodeSig sucked to write. I'm just going to leave that there. bigints in go are really hard because they are
// all pointers, which, not a bad thing necessarily, but it gets really confusing when you're porting code.
func (h *HeaderFields) decodeSig(s string, plen uint, e, n *big.Int) string {
	var i, dec int
	var plainChar string

	r, p, c := big.NewInt(0), big.NewInt(0), big.NewInt(0)
	for i = 0; i < len(s); i++ {
		dec = charMap(fmt.Sprintf("%c", s[i]))
		if dec < 0 {
			//h.Problems = append(h.Problems, errors.New("char decode out of range"))
		}
	}

	// this feels so wrong.
	r.Set(big.NewInt(int64(dec)))

	/*
	 let's be honest here, I'm not entirely sure I'm doing this right.
	 please don't ask me to explain it, I'm just porting the code.
	 https://git.io/vyrdU
	*/

	r.Mul(r, big.NewInt(int64(6*i)))
	c.Add(r, c)
	p.Exp(c, e, n)
	c.Set(big.NewInt(256))

	/*
		DivMod sets z to the quotient x div y and m to the modulus x mod y and returns the pair (z, m) for y != 0.
		If y == 0, a division-by-zero run-time panic occurs.
		https://git.io/vyrNu
	*/

	type localType struct {
		a int
	}

	for i := int(plen - 1); i >= 0; i-- {
		// I think this is right, the original C code is really confusing.
		p, r = p.DivMod(p, c, big.NewInt(0))
		plainChar += string(len(r.String()))
	}
	return plainChar
}

// charMap is a logic port from the libclamav code. it was faster/easier at the time to just port the logic than to
// look at making it better. there may be a better way.
func charMap(s string) int {
	for i := 0; i < 64; i++ {
		if refCharMap[i] == s {
			return i
		}
	}
	return -1
}

// readRadix implements bignum_fast.fp_read_radix() from tomsfastmath library with some minor changes since Go has a
// bigint/bignum type.
func readRadix(x *big.Int, s string, radix int) error {
	if radix < 2 || radix > 64 {
		return errors.New("invalid signature base")
	}

	var refChar string
	var y int

	// TODO find a better way to do this, this seems really convoluted.
	for i := 0; i < len(s); i++ {
		/*
			if the base is < 36, conversions are case-insensitive.
			therefore, 1AB == 1ab in hex.
		*/
		if radix <= 36 {
			refChar = strings.ToUpper(string(s[i]))
		} else {
			refChar = string(s[i])
		}
		for y = 0; y < 64; y++ {
			if refChar == string(refString[y]) {
				break
			}
		}

		/*
			if the character was found in the reference string
			and is less than the given base, add it to our number.
		*/
		if y < radix {
			x.Mul(x, x)
			x.Add(x, x)
		} else {
			break
		}
	}

	return nil
}
