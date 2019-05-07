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
	FAN_REPORT_FID      = 0x00000200
	FAN_MARK_ADD        = 0x00000001
	FAN_MARK_FILESYSTEM = 0x00000100
	FAN_ONDIR           = 0x40000000 /* event occurred against dir */
	FAN_MOVED_FROM      = 0x00000040 /* File was moved from X */
	FAN_MOVED_TO        = 0x00000080 /* File was moved to Y */
	FAN_CREATE          = 0x00000100 /* Subfile was created */
	FAN_DELETE          = 0x00000200 /* Subfile was deleted */
	FAN_DELETE_SELF     = 0x00000400 /* Self was deleted */
	FAN_MOVE_SELF       = 0x00000800 /* Self was moved */
	FAN_EVENT_ON_CHILD  = 0x08000000 /* interested in child events */
	AT_FDCWD            = -100
)
const markFlags = FAN_MARK_ADD | FAN_MARK_FILESYSTEM
const markMask = FAN_ONDIR | FAN_MOVED_FROM | FAN_MOVED_TO | FAN_CREATE | FAN_DELETE

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

type FileChange struct {
	FolderPath string
	ChangeType int
}

const (
	Creation = iota
	Deletion
)

func FanotifyInit(changeReceiver chan<- FileChange) {

	fan, err := unix.FanotifyInit(FAN_REPORT_FID, 0)
	if err != nil {
		fmt.Println(err)
		panic("could not call fanotifyinit")
	}

	err = unix.FanotifyMark(fan, markFlags, markMask, AT_FDCWD, "/")

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

		fd, err := unix.OpenByHandleAt(AT_FDCWD, unixFileHandle, 0)
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
		// 	fmt.Println("FAN_CREATE")
		// }
		// if meta.Mask&unix.IN_DELETE > 0 {
		// 	fmt.Println("FAN_DELETE")
		// }
		// if meta.Mask&unix.IN_DELETE_SELF > 0 {
		// 	fmt.Println("FAN_DELETE_SELF")
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
		// 	fmt.Println("FAN_MOVE_SELF")
		// }
		// if meta.Mask&unix.IN_MOVED_FROM > 0 {
		// 	fmt.Println("FAN_MOVED_FROM")
		// }
		// if meta.Mask&unix.IN_MOVED_TO > 0 {
		// 	fmt.Println("FAN_MOVED_TO")
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
