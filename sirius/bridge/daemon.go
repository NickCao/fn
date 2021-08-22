package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

var Endian = binary.LittleEndian

const WORKER_MAGIC_1 uint64 = 0x6e697863
const WORKER_MAGIC_2 uint64 = 0x6478696f
const PROTOCOL_VERSION uint64 = (1<<8 | 32)
const NAR_VERSION_MAGIC = "nix-archive-1"

func GET_PROTOCOL_MAJOR(x uint64) uint64 {
	return x & 0xff00
}

func GET_PROTOCOL_MINOR(x uint64) uint64 {
	return x & 0x00ff
}

type WorkerOp uint64

const (
	IsValidPath                 WorkerOp = 1
	HasSubstitutes                       = 3
	QueryReferrers                       = 6
	AddToStore                           = 7
	BuildPaths                           = 9
	EnsurePath                           = 10
	AddTempRoot                          = 11
	AddIndirectRoot                      = 12
	SyncWithGC                           = 13
	FindRoots                            = 14
	SetOptions                           = 19
	CollectGarbage                       = 20
	QuerySubstitutablePathInfo           = 21
	QueryAllValidPaths                   = 23
	QueryFailedPaths                     = 24
	ClearFailedPaths                     = 25
	QueryPathInfo                        = 26
	QueryPathFromHashPart                = 29
	QuerySubstitutablePathInfos          = 30
	QueryValidPaths                      = 31
	QuerySubstitutablePaths              = 32
	QueryValidDerivers                   = 33
	OptimiseStore                        = 34
	VerifyStore                          = 35
	BuildDerivation                      = 36
	AddSignatures                        = 37
	NarFromPath                          = 38
	AddToStoreNar                        = 39
	QueryMissing                         = 40
	QueryDerivationOutputMap             = 41
	RegisterDrvOutput                    = 42
	QueryRealisation                     = 43
	AddMultipleToStore                   = 44
)

type Daemon struct {
}

func (d *Daemon) ProcessConn(conn io.Reader) error {
	var magic, version, affinity, padding uint64
	err := binary.Read(conn, Endian, &magic)
	if err != nil {
		return err
	}
	if magic != WORKER_MAGIC_1 {
		return fmt.Errorf("protocol mismatch")
	}
	err = binary.Read(conn, Endian, &version)
	if err != nil {
		return err
	}
	if version < 0x10a {
		return fmt.Errorf("client too old")
	}
	if GET_PROTOCOL_MINOR(version) >= 14 {
		err = binary.Read(conn, Endian, &padding)
		if err != nil {
			return err
		}
		err = binary.Read(conn, Endian, &affinity)
		if err != nil {
			return err
		}
		fmt.Printf("set affinity to: %d\n", affinity)
		// TODO: set affinity
	}
	/*
		err = binary.Read(conn, Endian, &padding)
		if err != nil {
			return err
		}
	*/
	fmt.Printf("start handling ops: client version %d %d\n", GET_PROTOCOL_MAJOR(version), GET_PROTOCOL_MINOR(version))
	var op uint64
	for {
		err = binary.Read(conn, Endian, &op)
		if err != nil {
			return err
		}
		fmt.Printf("get op: %d\n", op)
		switch WorkerOp(op) {
		case IsValidPath:
		case HasSubstitutes:
		case QueryReferrers:
		case AddToStore:
		case BuildPaths:
		case EnsurePath:
		case AddTempRoot:
		case AddIndirectRoot:
		case SyncWithGC:
		case FindRoots:
		case SetOptions:
		case CollectGarbage:
		case QuerySubstitutablePathInfo:
		case QueryAllValidPaths:
		case QueryFailedPaths:
		case ClearFailedPaths:
		case QueryPathInfo:
		case QueryPathFromHashPart:
		case QuerySubstitutablePathInfos:
		case QueryValidPaths:
			var numPaths uint64
			err = binary.Read(conn, Endian, &numPaths)
			if err != nil {
				return err
			}
			for i := uint64(0); i < numPaths; i++ {
				path, err := readString(conn)
				if err != nil {
					return err
				}
				fmt.Printf("query path: %s\n", path)
			}
			var substitute uint64 = 0
			if GET_PROTOCOL_MINOR(version) >= 27 {
				err = binary.Read(conn, Endian, &substitute)
				fmt.Printf("whether to substitute: %d\n", substitute)
			}
		case QuerySubstitutablePaths:
		case QueryValidDerivers:
		case OptimiseStore:
		case VerifyStore:
		case BuildDerivation:
		case AddSignatures:
		case NarFromPath:
		case AddToStoreNar:
		case QueryMissing:
		case QueryDerivationOutputMap:
		case RegisterDrvOutput:
		case QueryRealisation:
		case AddMultipleToStore:
			var repair, dontCheckSigs uint64
			err = binary.Read(conn, Endian, &repair)
			if err != nil {
				return err
			}
			err = binary.Read(conn, Endian, &dontCheckSigs)
			if err != nil {
				return err
			}
			fr := NewFramedReader(conn)
			var numPaths uint64
			err = binary.Read(fr, Endian, &numPaths)
			if err != nil {
				return err
			}
			fmt.Printf("storing %d pathes\n", numPaths)
			for i := uint64(0); i < numPaths; i++ {
				path, err := readString(fr)
				if err != nil {
					return err
				}
				deriver, err := readString(fr)
				if err != nil {
					return err
				}
				hash, err := readString(fr)
				if err != nil {
					return err
				}
				fmt.Printf("path info: %s %s %s\n", path, deriver, hash)
				var numRefs uint64
				err = binary.Read(fr, Endian, &numRefs)
				if err != nil {
					return err
				}
				for j := uint64(0); j < numRefs; j++ {
					ref, err := readString(fr)
					if err != nil {
						return err
					}
					fmt.Printf("path ref: %s\n", ref)
				}
				var regTime, narSize uint64
				err = binary.Read(fr, Endian, &regTime)
				if err != nil {
					return err
				}
				err = binary.Read(fr, Endian, &narSize)
				if err != nil {
					return err
				}
				var ultimate uint64
				err = binary.Read(fr, Endian, &ultimate)
				if err != nil {
					return err
				}
				fmt.Printf("regtime: %d, narsize: %d, ultimate %d\n", regTime, narSize, ultimate)
				var numSigs uint64
				err = binary.Read(fr, Endian, &numSigs)
				if err != nil {
					return err
				}
				for k := uint64(0); k < numSigs; k++ {
					sig, err := readString(fr)
					if err != nil {
						return err
					}
					fmt.Printf("sig: %s\n", sig)
				}
				ca, err := readString(fr)
				if err != nil {
					return err
				}
				fmt.Printf("ca: %s\n", ca)
				vmagic, err := readString(fr)
				if err != nil {
					return err
				}
				if vmagic != NAR_VERSION_MAGIC {
					return fmt.Errorf("unlikely magic mismatch")
				}
				fmt.Println(readArchive(fr, ""))
			}
		default:
			return fmt.Errorf("invalid op")
		}
	}
}

type tp int

const (
	tpUnknown tp = iota
	tpRegular
	tpDirectory
	tpSymlink
)

func readArchive(conn io.Reader, path string) error {
	s1, err := readString(conn)
	if err != nil {
		return err
	}
	if s1 != "(" {
		return fmt.Errorf("expected open tag")
	}
	ctp := tpUnknown
outer:
	for {
		s2, err := readString(conn)
		if err != nil {
			return err
		}
		fmt.Printf("s2: %s\n", s2)
		switch s2 {
		case ")":
			break outer
		case "type":
			if ctp != tpUnknown {
				return fmt.Errorf("multiple type field")
			}
			t, err := readString(conn)
			if err != nil {
				return err
			}
			switch t {
			case "regular":
				ctp = tpRegular
			case "directory":
				ctp = tpDirectory
			case "symlink":
				ctp = tpSymlink
			default:
				return fmt.Errorf("invalid type field")
			}
		case "contents":
			if ctp != tpRegular {
				return fmt.Errorf("bad archive")
			}
			var size, sizePadded uint64
			err = binary.Read(conn, Endian, &size)
			if err != nil {
				return err
			}
			if size%8 == 0 {
				sizePadded = size
			} else {
				sizePadded = size + 8 - (size % 8)
			}
			_, err = io.CopyN(io.Discard, conn, int64(sizePadded))
			if err != nil {
				return err
			}
			fmt.Printf("content at: %s\n", path)
		case "executable":
			if ctp != tpRegular {
				return fmt.Errorf("bad archive")
			}
			marker, err := readString(conn)
			if err != nil {
				return err
			}
			if marker != "" {
				return fmt.Errorf("non empty x marker")
			}
		case "entry":
			if ctp != tpDirectory {
				return fmt.Errorf("bad archive")
			}
			var prevname, name string
			s3, err := readString(conn)
			if err != nil {
				return err
			}
			if s3 != "(" {
				return fmt.Errorf("expected open tag")
			}
		inner:
			for {
				s4, err := readString(conn)
				if err != nil {
					return err
				}
				switch s4 {
				case ")":
					break inner
				case "name":
					name, err = readString(conn)
					if err != nil {
						return err
					}
					if name == "" || name == "." || name == ".." || strings.ContainsRune(name, '/') || strings.ContainsRune(name, 0) {
						return fmt.Errorf("invalid name")
					}
					if name <= prevname {
						return fmt.Errorf("name not sorted")
					}
					prevname = name
				case "node":
					if name == "" {
						return fmt.Errorf("name missing")
					}
					readArchive(conn, path+"/"+name)
				default:
					return fmt.Errorf("bad archive")
				}
			}
		case "target":
			if ctp != tpSymlink {
				return fmt.Errorf("bad archive")
			}
			_, err := readString(conn)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("bad archive")
		}
	}
	return nil
}

func readString(conn io.Reader) (string, error) {
	var lenPath uint64
	err := binary.Read(conn, Endian, &lenPath)
	if err != nil {
		return "", err
	}
	var lenPathPadded uint64
	if lenPath%8 == 0 {
		lenPathPadded = lenPath
	} else {
		lenPathPadded = lenPath + 8 - (lenPath % 8)
	}
	buf := make([]byte, lenPathPadded)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
	if uint64(n) != lenPathPadded {
		return "", fmt.Errorf("unlikely readstring fail")
	}
	return string(buf[:lenPath]), nil
}

type FramedReader struct {
	rd  io.Reader
	buf *bytes.Buffer
}

func (f *FramedReader) Read(p []byte) (int, error) {
	if f.buf.Len() < len(p) {
		var size uint64
		err := binary.Read(f.rd, Endian, &size)
		if err != nil {
			return 0, err
		}
		buf := make([]byte, size)
		_, err = io.ReadFull(f.rd, buf)
		if err != nil {
			return 0, err
		}
		n, err := f.buf.Write(buf)
		if err != nil {
			return 0, err
		}
		if uint64(n) != size {
			return 0, fmt.Errorf("unlikely write frame fail")
		}
		//fmt.Printf("read frame of size: %d\n", size)
	}
	return f.buf.Read(p)
}

func NewFramedReader(rd io.Reader) *FramedReader {
	return &FramedReader{rd: rd, buf: bytes.NewBuffer(nil)}
}
