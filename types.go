package main

import (
	"time"
)

type ClamAV struct {
	Header     HeaderFields
	Definition AVDefinition
	Problems   map[string]error
}

type HeaderFields struct {
	CreationTime  time.Time
	Version       uint
	Signatures    uint
	Functionality uint
	MD5Hash       string
	MD5Valid      bool
	DSignature    string
	DSigValid     bool
	Builder       string
	Stime         uint
	Problems      []error
}

type AVDefinition struct {
	Body []byte
}
