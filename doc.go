/*

Raw file header example:

ClamAV-VDB:07 Mar 2017 08-02 -0500:23182:1741572:63:c1537143239006af01e814a4dcd58a48:QC2ZncCPK0AzfYPW8OKvde9GFOO1HyH5qbozl9JZbmlOmZnSV55zWaP9yH9tXiS+JmZWA1277X6pBeTHPCcaqUDakke4W58duZ5mavDGJoWekl3q/5RgVeAg39cM1X4zNf6gER8G+HIWDUka0sRQWal1KXAb1UWkFoKsbHVqgVi:neo:1488891746

Field Definitions:
ClamAV-VDB: Header, defines file type.
07 Mar 2017 08-02 -0500: Creation time.
23182: Version Number
1741572: Number of signatures.
63: Functionality level.
c1...58a48: MD5 checksum.
QC...VqgVi: Digital Signature. Type Unknown.
neo: Builder Name.
Creation time in Epoch Seconds: old file format.

Actual definition:
struct cl_cvd {
char *time;		    2
unsigned int version;   3
unsigned int sigs;	    4
unsigned int fl;	    5
// padding
char *md5;		    /6
char *dsig;		    7
char *builder;	    8
unsigned int stime;	    9
};

*/
package main
