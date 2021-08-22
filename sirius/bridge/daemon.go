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

// TODO: properly handle stderr
const STDERR_LAST uint64 = 0x616c7473

func GET_PROTOCOL_MAJOR(x uint64) uint64 {
	return x & 0xff00
}

func GET_PROTOCOL_MINOR(x uint64) uint64 {
	return x & 0x00ff
}

type BuildStatus uint64

const (
	Built BuildStatus = iota
	Substituted
	AlreadyValid
	PermanentFailure
	InputRejected
	OutputRejected
	TransientFailure // possibly transient
	CachedFailure    // no longer used
	TimedOut
	MiscFailure
	DependencyFailed
	LogLimitExceeded
	NotDeterministic
)

type WorkerOp uint64

const (
	Nop                         WorkerOp = 0
	IsValidPath                          = 1
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

func (d *Daemon) ProcessConn(conn io.ReadWriter) error {
	magic, err := readUInt64(conn)
	if err != nil {
		return err
	}
	if magic != WORKER_MAGIC_1 {
		return fmt.Errorf("protocol mismatch")
	}

	err = writeUInt64(conn, WORKER_MAGIC_2)
	if err != nil {
		return err
	}
	err = writeUInt64(conn, PROTOCOL_VERSION)
	if err != nil {
		return err
	}

	version, err := readUInt64(conn)
	if err != nil {
		return err
	}
	if version < 0x10a {
		return fmt.Errorf("client too old")
	}

	if GET_PROTOCOL_MINOR(version) >= 14 {
		doAff, err := readUInt64(conn)
		if err != nil {
			return err
		}
		if doAff == 0 {
			affinity, err := readUInt64(conn)
			if err != nil {
				return err
			}
			fmt.Printf("set affinity to: %d\n", affinity)
		}
	}
	fmt.Printf("start handling ops: client version %d %d\n", GET_PROTOCOL_MAJOR(version), GET_PROTOCOL_MINOR(version))

	err = writeUInt64(conn, STDERR_LAST)
	if err != nil {
		return err
	}
	//f, _ := os.OpenFile("/tmp/proto" ,os.O_RDWR|os.O_CREATE, 0755)
	//go io.Copy(conn, io.TeeReader(backend, f))
	for {
		op, err := readUInt64(conn)
		if err != nil {
			return err
		}
		fmt.Printf("op: %d\n", op)
		switch WorkerOp(op) {
		case Nop:
		case IsValidPath:
			return fmt.Errorf("not implemented")
		case HasSubstitutes:
			return fmt.Errorf("not implemented")
		case QueryReferrers:
			return fmt.Errorf("not implemented")
		case AddToStore:
			return fmt.Errorf("not implemented")
		case BuildPaths:
			return fmt.Errorf("not implemented")
		case EnsurePath:
			return fmt.Errorf("not implemented")
		case AddTempRoot:
			return fmt.Errorf("not implemented")
		case AddIndirectRoot:
			return fmt.Errorf("not implemented")
		case SyncWithGC:
			return fmt.Errorf("not implemented")
		case FindRoots:
			return fmt.Errorf("not implemented")
		case SetOptions:
			return fmt.Errorf("not implemented")
		case CollectGarbage:
			return fmt.Errorf("not implemented")
		case QuerySubstitutablePathInfo:
			return fmt.Errorf("not implemented")
		case QueryAllValidPaths:
			return fmt.Errorf("not implemented")
		case QueryPathInfo:
			if GET_PROTOCOL_MINOR(version) < 17 {
				return fmt.Errorf("unsupported client version")
			}
			path, err := readString(conn)
			if err != nil {
				return err
			}
			fmt.Printf("QueryPathInfo: path %s\n", path)
			err = writeUInt64(conn, STDERR_LAST)
			if err != nil {
				return err
			}
			// TODO: fake that the path exists
			err = writeUInt64(conn, 1)
			if err != nil {
				return err
			}
			err = writeString(conn, "") // deriver
			if err != nil {
				return err
			}
			err = writeString(conn, "0sg9f58l1jj88w6pdrfdpj5x9b1zrwszk84j81zvby36q9whhhqa") // narhash
			if err != nil {
				return err
			}
			err = writeUInt64(conn, 0) // TODO: refs
			if err != nil {
				return err
			}
			err = writeUInt64(conn, 0) // regtime
			if err != nil {
				return err
			}
			err = writeUInt64(conn, 120) // nar size
			if err != nil {
				return err
			}
			err = writeUInt64(conn, 0) // ultimate
			if err != nil {
				return err
			}
			err = writeUInt64(conn, 0) // TODO: sigs
			if err != nil {
				return err
			}
			err = writeString(conn, "") // TODO: ca
			if err != nil {
				return err
			}
		case QueryPathFromHashPart:
			return fmt.Errorf("not implemented")
		case QuerySubstitutablePathInfos:
			return fmt.Errorf("not implemented")
		case QueryValidPaths:
			paths, err := readStrings(conn)
			if err != nil {
				return err
			}
			var substitute uint64 = 0
			if GET_PROTOCOL_MINOR(version) >= 27 {
				substitute, err = readUInt64(conn)
				if err != nil {
					return err
				}
			}
			fmt.Printf("QueryValidPaths: paths: %s, substitute: %d\n", paths, substitute)
			err = writeUInt64(conn, STDERR_LAST)
			if err != nil {
				return err
			}
			// TODO: impl path query
			err = writeUInt64(conn, 0)
			if err != nil {
				return err
			}
		case QuerySubstitutablePaths:
			return fmt.Errorf("not implemented")
		case QueryValidDerivers:
			return fmt.Errorf("not implemented")
		case OptimiseStore:
			return fmt.Errorf("not implemented")
		case VerifyStore:
			return fmt.Errorf("not implemented")
		case BuildDerivation:
			path, err := readString(conn)
			if err != nil {
				return err
			}
			fmt.Printf("build derivation: %s\n", path)
			var nr uint64
			err = binary.Read(conn, Endian, &nr)
			if err != nil {
				return err
			}
			for n := uint64(0); n < nr; n++ {
				name, err := readString(conn)
				if err != nil {
					return err
				}
				pathS, err := readString(conn)
				if err != nil {
					return err
				}
				hashAlgo, err := readString(conn)
				if err != nil {
					return err
				}
				hash, err := readString(conn)
				if err != nil {
					return err
				}
				fmt.Printf("output: %s %s %s %s\n", name, pathS, hashAlgo, hash)
			}
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
				fmt.Printf("input src path: %s\n", path)
			}
			platform, err := readString(conn)
			if err != nil {
				return err
			}
			builder, err := readString(conn)
			if err != nil {
				return err
			}
			fmt.Printf("platform: %s, builder: %s\n", platform, builder)
			var numArgs uint64
			err = binary.Read(conn, Endian, &numArgs)
			if err != nil {
				return err
			}
			for i := uint64(0); i < numArgs; i++ {
				arg, err := readString(conn)
				if err != nil {
					return err
				}
				fmt.Printf("arg: %s\n", arg)
			}
			var numEnvs uint64
			err = binary.Read(conn, Endian, &numEnvs)
			if err != nil {
				return err
			}
			for i := uint64(0); i < numEnvs; i++ {
				key, err := readString(conn)
				if err != nil {
					return err
				}
				value, err := readString(conn)
				if err != nil {
					return err
				}
				fmt.Printf("env: %s %s\n", key, value)
			}
			var buildMode uint64
			err = binary.Read(conn, Endian, &buildMode)
			if err != nil {
				return err
			}
			fmt.Printf("build mode: %d\n", buildMode)
			err = writeUInt64(conn, STDERR_LAST)
			if err != nil {
				return err
			}
			err = writeUInt64(conn, uint64(Built))
			if err != nil {
				return err
			}
			err = writeString(conn, "built")
			if err != nil {
				return err
			}
			if GET_PROTOCOL_MINOR(version) >= 29 {
				var timesBuilt, isNonDeterministic, startTime, stopTime uint64
				err = writeUInt64(conn, timesBuilt)
				if err != nil {
					return err
				}
				err = writeUInt64(conn, isNonDeterministic)
				if err != nil {
					return err
				}
				err = writeUInt64(conn, startTime)
				if err != nil {
					return err
				}
				err = writeUInt64(conn, stopTime)
				if err != nil {
					return err
				}
			}
			if GET_PROTOCOL_MINOR(version) >= 28 {
				// TODO: write drv outputs map
				err = writeUInt64(conn, 0)
				if err != nil {
					return err
				}
			}
		case AddSignatures:
			return fmt.Errorf("not implemented")
		case NarFromPath:
			path, err := readString(conn)
			if err != nil {
				return err
			}
			fmt.Printf("NarFromPath: %s\n", path)
			err = writeUInt64(conn, STDERR_LAST)
			if err != nil {
				return err
			}
			err = writeString(conn, NAR_VERSION_MAGIC)
			if err != nil {
				return err
			}
			err = writeString(conn, "(")
			if err != nil {
				return err
			}
			err = writeString(conn, "type")
			if err != nil {
				return err
			}
			err = writeString(conn, "regular")
			if err != nil {
				return err
			}
			err = writeString(conn, "contents")
			if err != nil {
				return err
			}
			err = writeString(conn, "hello")
			if err != nil {
				return err
			}
			err = writeString(conn, ")")
			if err != nil {
				return err
			}
		case AddToStoreNar:
			return fmt.Errorf("not implemented")
		case QueryMissing:
			return fmt.Errorf("not implemented")
		case QueryDerivationOutputMap:
			return fmt.Errorf("not implemented")
		case RegisterDrvOutput:
			return fmt.Errorf("not implemented")
		case QueryRealisation:
			return fmt.Errorf("not implemented")
		case AddMultipleToStore:
			repair, err := readUInt64(conn)
			if err != nil {
				return err
			}
			dontCheckSigs, err := readUInt64(conn)
			if err != nil {
				return err
			}
			fmt.Printf("AddMultipleToStore: repair: %d, dontCheckSigs %d\n", repair, dontCheckSigs)
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
			err = writeUInt64(conn, STDERR_LAST)
			if err != nil {
				return err
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
		case "target":
			if ctp != tpSymlink {
				return fmt.Errorf("bad archive")
			}
			_, err := readString(conn)
			if err != nil {
				return err
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
					err = readArchive(conn, path+"/"+name)
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf("bad archive")
				}
			}
		default:
			return fmt.Errorf("bad archive")
		}
	}
	return nil
}

func roundPadding(n uint64) uint64 {
	rem := n % 8
	if rem == 0 {
		return n
	}
	return n - rem + 8
}

func readUInt64(conn io.Reader) (uint64, error) {
	var num uint64
	err := binary.Read(conn, Endian, &num)
	return num, err
}

func writeUInt64(conn io.Writer, num uint64) error {
	err := binary.Write(conn, Endian, &num)
	return err
}

func writeString(conn io.Writer, s string) error {
	buf := []byte(s)
	sLen := uint64(len(buf))
	buf = append(buf, make([]byte, roundPadding(sLen)-sLen)...)
	err := writeUInt64(conn, sLen)
	if err != nil {
		return err
	}
	_, err = conn.Write(buf)
	return err
}

func readString(conn io.Reader) (string, error) {
	lenPath, err := readUInt64(conn)
	if err != nil {
		return "", err
	}
	buf := make([]byte, roundPadding(lenPath))
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return "", err
	}
	return string(buf[:lenPath]), nil
}

func readStrings(conn io.Reader) ([]string, error) {
	n, err := readUInt64(conn)
	if err != nil {
		return nil, err
	}
	ss := make([]string, n)
	for i := uint64(0); i < n; i++ {
		s, err := readString(conn)
		if err != nil {
			return nil, err
		}
		ss[i] = s
	}
	return ss, nil
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
