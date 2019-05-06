package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"unsafe"

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
const MARK_FLAGS = FAN_MARK_ADD | FAN_MARK_FILESYSTEM
const MARK_MASK = FAN_ONDIR | FAN_MOVED_FROM | FAN_MOVED_TO | FAN_CREATE | FAN_DELETE

type FanotifyInfoHeader struct {
	InfoType uint8
	Pad      uint8
	Len      uint16
}

type FileHandle struct {
	HandleBytes uint32
	HandleType  int32
	// file indentiefier of arbitrary length
}

type FanotifyEventFid struct {
	KernelFsidT [2]int32
	FileHandle  FileHandle
}

type FanotifyEventInfoFid struct {
	Hdr      FanotifyInfoHeader
	EventFid FanotifyEventFid
	/*
	 * Following is an opaque struct file_handle that can be passed as
	 * an argument to open_by_handle_at(2).
	 */
	// unsigned char handle[0];
}

func fanotifyInit() error {
	fd, err := unix.FanotifyInit(FAN_REPORT_FID, 0)
	if err != nil {
		return err
	}
	err = unix.FanotifyMark(fd, MARK_FLAGS, MARK_MASK, AT_FDCWD, "")

	if err != nil {
		return err
	}

	f := os.NewFile(uintptr(fd), "")
	r := bufio.NewReader(f)

	metaBuff := make([]byte, 24)

	for {
		n, err := r.Read(metaBuff)
		if err != nil {
			return err
		}

		if n < 0 || n > 24 {
			return errors.New(fmt.Sprintf("read %d bytes", n))
		}

		meta := *((*unix.FanotifyEventMetadata)(unsafe.Pointer(&metaBuff[0])))
		bytesLeft := int(meta.Event_len - uint32(meta.Metadata_len))
		infoBuff := make([]byte, bytesLeft)
		n, err = r.Read(infoBuff)
		if err != nil {
			return err
		}

		if n < 0 || n > bytesLeft {
			return errors.New(fmt.Sprintf("read %d bytes", n))
		}

		info := *((*FanotifyEventInfoFid)(unsafe.Pointer(&infoBuff[0])))

		if info.Hdr.InfoType != 1 { // TODO: properly define constant
			continue
		}

		handleStart := uint32(unsafe.Sizeof(info))
		handleLen := info.EventFid.FileHandle.HandleBytes
		handleBytes := infoBuff[handleStart : handleStart+handleLen]
		unixFileHandle := unix.NewFileHandle(info.EventFid.FileHandle.HandleType, handleBytes)

		// fmt.Printf("info: %+v\n", info)
		fd, err = unix.OpenByHandleAt(AT_FDCWD, unixFileHandle, 0)
		if err != nil {
			// fmt.Println("openbyhandleat")
			fmt.Println(err)
			continue
			// return err
		}

		sym := fmt.Sprintf("/proc/self/fd/%d", fd)
		name := make([]byte, 200)
		n, err = unix.Readlink(sym, name)
		if err != nil {
			fmt.Println("unix.readlink")
			return err
		}

		fmt.Println(string(name))
		fmt.Printf("meta: %+v\n", meta)
		fmt.Printf("info: %+v\n", info)
		fmt.Printf("metaBuff: %v\n", metaBuff)
		fmt.Printf("infoBuff: %v\n", infoBuff)
		// fmt.Printf("cutOff: %v\n", metaBuff[handleStart+handleLen:])

		if meta.Mask&unix.IN_ACCESS > 0 {
			fmt.Println("FAN_ACCESS")
		}
		if meta.Mask&unix.IN_ATTRIB > 0 {
			fmt.Println("FAN_ATTRIB")
		}
		if meta.Mask&unix.IN_CLOSE_NOWRITE > 0 {
			fmt.Println("FAN_CLOSE_NOWRITE")
		}
		if meta.Mask&unix.IN_CLOSE_WRITE > 0 {
			fmt.Println("FAN_CLOSE_WRITE")
		}
		if meta.Mask&unix.IN_CREATE > 0 {
			fmt.Println("FAN_CREATE")
		}
		if meta.Mask&unix.IN_DELETE > 0 {
			fmt.Println("FAN_DELETE")
		}
		if meta.Mask&unix.IN_DELETE_SELF > 0 {
			fmt.Println("FAN_DELETE_SELF")
		}
		if meta.Mask&unix.IN_IGNORED > 0 {
			fmt.Println("FAN_IGNORED")
		}
		if meta.Mask&unix.IN_ISDIR > 0 {
			fmt.Println("FAN_ISDIR")
		}
		if meta.Mask&unix.IN_MODIFY > 0 {
			fmt.Println("FAN_MODIFY")
		}
		if meta.Mask&unix.IN_MOVE_SELF > 0 {
			fmt.Println("FAN_MOVE_SELF")
		}
		if meta.Mask&unix.IN_MOVED_FROM > 0 {
			fmt.Println("FAN_MOVED_FROM")
		}
		if meta.Mask&unix.IN_MOVED_TO > 0 {
			fmt.Println("FAN_MOVED_TO")
		}
		if meta.Mask&unix.IN_OPEN > 0 {
			fmt.Println("FAN_OPEN")
		}
		if meta.Mask&unix.IN_Q_OVERFLOW > 0 {
			fmt.Println("FAN_Q_OVERFLOW")
		}
		if meta.Mask&unix.IN_UNMOUNT > 0 {
			fmt.Println("FAN_UNMOUNT")
		}
	}

	return nil
}
