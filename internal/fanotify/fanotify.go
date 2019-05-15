package fanotify

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"unsafe"

	"github.com/ozeidan/gosearch/internal/config"
	"golang.org/x/sys/unix"
)

const (
	fanReportFid      = 0x00000200
	fanMarkAdd        = 0x00000001
	fanMarkFilesystem = 0x00000100
	fanOndir          = 0x40000000 /* event occurred against dir */
	fanMovedFrom      = 0x00000040 /* File was moved from X */
	fanMovedTo        = 0x00000080 /* File was moved to Y */
	fanCreate         = 0x00000100 /* Subfile was created */
	fanDelete         = 0x00000200 /* Subfile was deleted */
	fanDeleteSelf     = 0x00000400 /* Self was deleted */
	fanMoveSelf       = 0x00000800 /* Self was moved */
	fanEventOnChild   = 0x08000000 /* interested in child events */
	atFDCWD           = -100
)
const markFlags = fanMarkAdd | fanMarkFilesystem
const markMask = fanOndir | fanMovedFrom | fanMovedTo | fanCreate | fanDelete

type fanotifyInfoHeader struct {
	infoType uint8
	pad      uint8
	Len      uint16
}

type fileHandle struct {
	handleBytes uint32
	handleType  int32
	// file indentiefier of arbitrary length
}

type fanotifyEventFid struct {
	kernelFsidT [2]int32
	fileHandle  fileHandle
}

type fanotifyEventInfoFid struct {
	hdr      fanotifyInfoHeader
	eventFid fanotifyEventFid
}

// FileChange describes the event of changes in a directory
// FolderPath is the path of the directory
// Changetype is either Creation or Deletion
type FileChange struct {
	FolderPath string
	ChangeType int
}

const (
	// Creation of a file/directory
	Creation = iota
	// Deletion of a file/directory
	Deletion
)

// Listen starts listening for created/deleted/moved
// files in the whole file system
// changeReceiver is a channel that FileChange structs,
// which describe the events, will be sent through
func Listen(changeReceiver chan<- FileChange) {
	log.Println("starting to listen on fanotify events")
	fan, err := unix.FanotifyInit(fanReportFid, 0)
	if err != nil {
		fmt.Println(err)
		panic("could not call fanotifyinit")
	}

	err = unix.FanotifyMark(fan, markFlags, markMask, atFDCWD, "/")

	if err != nil {
		fmt.Println(err)
		panic("could not call fanotifymark")
	}

	log.Println("fanotify initialized")

	f := os.NewFile(uintptr(fan), "")
	r := bufio.NewReader(f)

	metaBuff := make([]byte, 24)

	for {
		n, err := r.Read(metaBuff)
		if err != nil {
			continue
		}

		if n < 0 || n > 24 {
			continue
		}

		meta := *((*unix.FanotifyEventMetadata)(unsafe.Pointer(&metaBuff[0])))
		bytesLeft := int(meta.Event_len - uint32(meta.Metadata_len))
		infoBuff := make([]byte, bytesLeft)
		n, err = r.Read(infoBuff)
		if err != nil {
			continue
		}

		if n < 0 || n > bytesLeft {
			continue
		}

		info := *((*fanotifyEventInfoFid)(unsafe.Pointer(&infoBuff[0])))

		if info.hdr.infoType != 1 { // TODO: properly define constant
			continue
		}

		handleStart := uint32(unsafe.Sizeof(info))
		handleLen := info.eventFid.fileHandle.handleBytes
		handleBytes := infoBuff[handleStart : handleStart+handleLen]
		unixFileHandle := unix.NewFileHandle(info.eventFid.fileHandle.handleType, handleBytes)

		fd, err := unix.OpenByHandleAt(atFDCWD, unixFileHandle, 0)
		if err != nil {
			log.Println("could not call OpenByHandleAt:", err)
			continue
		}

		sym := fmt.Sprintf("/proc/self/fd/%d", fd)
		path := make([]byte, 200)
		pathLength, err := unix.Readlink(sym, path)

		if err != nil {
			log.Println("could not call Readlink:", err)
			continue
		}
		path = path[:pathLength]
		log.Println("received event, path:", string(path),
			"flags:", maskToString(meta.Mask))
		if config.IsPathFiltered(string(path)) {
			continue
		}

		changeType := 0
		if meta.Mask&unix.IN_CREATE > 0 ||
			meta.Mask&unix.IN_MOVED_TO > 0 {
			changeType = Creation
		}
		if meta.Mask&unix.IN_DELETE > 0 ||
			meta.Mask&unix.IN_MOVED_FROM > 0 {
			changeType = Deletion
		}

		change := FileChange{
			string(path),
			changeType,
		}
		changeReceiver <- change
	}
}

func maskToString(mask uint64) string {
	var flags []string
	if mask&unix.IN_ACCESS > 0 {
		flags = append(flags, "FAN_ACCESS")
	}
	if mask&unix.IN_ATTRIB > 0 {
		flags = append(flags, "FAN_ATTRIB")
	}
	if mask&unix.IN_CLOSE_NOWRITE > 0 {
		flags = append(flags, "FAN_CLOSE_NOWRITE")
	}
	if mask&unix.IN_CLOSE_WRITE > 0 {
		flags = append(flags, "FAN_CLOSE_WRITE")
	}
	if mask&unix.IN_CREATE > 0 {
		flags = append(flags, "FAN_CREATE")
	}
	if mask&unix.IN_DELETE > 0 {
		flags = append(flags, "FAN_DELETE")
	}
	if mask&unix.IN_DELETE_SELF > 0 {
		flags = append(flags, "FAN_DELETE_SELF")
	}
	if mask&unix.IN_IGNORED > 0 {
		flags = append(flags, "FAN_IGNORED")
	}
	if mask&unix.IN_ISDIR > 0 {
		flags = append(flags, "FAN_ISDIR")
	}
	if mask&unix.IN_MODIFY > 0 {
		flags = append(flags, "FAN_MODIFY")
	}
	if mask&unix.IN_MOVE_SELF > 0 {
		flags = append(flags, "fanMoveSelf")
	}
	if mask&unix.IN_MOVED_FROM > 0 {
		flags = append(flags, "fanMovedFrom")
	}
	if mask&unix.IN_MOVED_TO > 0 {
		flags = append(flags, "fanMovedTo")
	}
	if mask&unix.IN_OPEN > 0 {
		flags = append(flags, "FAN_OPEN")
	}
	if mask&unix.IN_Q_OVERFLOW > 0 {
		flags = append(flags, "FAN_Q_OVERFLOW")
	}
	if mask&unix.IN_UNMOUNT > 0 {
		flags = append(flags, "FAN_UNMOUNT")
	}
	return strings.Join(flags, ", ")
}
