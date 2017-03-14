package main

import (
	"time"
)

// ClamAV is the typed structure of a ClamAV definition file.
type ClamAV struct {
	Header     HeaderFields
	Definition AVDefinition
	Problems   []error
}

// HeaderFields are the parsed 512 bytes of the antivirus definition.
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

// AVDefinition is the binary blob of antivirus definition data.
type AVDefinition struct {
	Body []byte
}
