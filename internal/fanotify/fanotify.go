package fanotify

import (
	"bufio"
	"fmt"
	"os"
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
			fmt.Println("could not call OpenByHandleAt:", err)
			continue
		}

		sym := fmt.Sprintf("/proc/self/fd/%d", fd)
		path := make([]byte, 200)
		pathLength, err := unix.Readlink(sym, path)
		if err != nil {
			fmt.Println("could not call Readlink:", err)
			continue
		}
		if config.IsPathFiltered(string(path)) {
			continue
		}

		// if meta.Mask&unix.IN_ACCESS > 0 {
		// 	fmt.Println("FAN_ACCESS")
		// }
		// if meta.Mask&unix.IN_ATTRIB > 0 {
		// 	fmt.Println("FAN_ATTRIB")
		// }
		// if meta.Mask&unix.IN_CLOSE_NOWRITE > 0 {
		// 	fmt.Println("FAN_CLOSE_NOWRITE")
		// }
		// if meta.Mask&unix.IN_CLOSE_WRITE > 0 {
		// 	fmt.Println("FAN_CLOSE_WRITE")
		// }
		// if meta.Mask&unix.IN_CREATE > 0 {
		// 	fmt.Println("fanCreate")
		// }
		// if meta.Mask&unix.IN_DELETE > 0 {
		// 	fmt.Println("fanDelete")
		// }
		// if meta.Mask&unix.IN_DELETE_SELF > 0 {
		// 	fmt.Println("fanDelete_SELF")
		// }
		// if meta.Mask&unix.IN_IGNORED > 0 {
		// 	fmt.Println("FAN_IGNORED")
		// }
		// if meta.Mask&unix.IN_ISDIR > 0 {
		// 	fmt.Println("FAN_ISDIR")
		// }
		// if meta.Mask&unix.IN_MODIFY > 0 {
		// 	fmt.Println("FAN_MODIFY")
		// }
		// if meta.Mask&unix.IN_MOVE_SELF > 0 {
		// 	fmt.Println("fanMoveSelf")
		// }
		// if meta.Mask&unix.IN_MOVED_FROM > 0 {
		// 	fmt.Println("fanMovedFrom")
		// }
		// if meta.Mask&unix.IN_MOVED_TO > 0 {
		// 	fmt.Println("fanMovedTo")
		// }
		// if meta.Mask&unix.IN_OPEN > 0 {
		// 	fmt.Println("FAN_OPEN")
		// }
		// if meta.Mask&unix.IN_Q_OVERFLOW > 0 {
		// 	fmt.Println("FAN_Q_OVERFLOW")
		// }
		// if meta.Mask&unix.IN_UNMOUNT > 0 {
		// 	fmt.Println("FAN_UNMOUNT")
		// }

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
			string(path[:pathLength]),
			changeType,
		}
		changeReceiver <- change
	}
}
