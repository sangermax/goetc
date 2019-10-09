package main

/*
#cgo CFLAGS: -I./
#cgo LDFLAGS: -L./so -lICReader -ldl
#include <stdio.h>
#include <stdlib.h>
#include "readerInf.h"
*/
import "C"
import (
	"unsafe"
)

type ReaderInf struct {
}

func (p *ReaderInf) FuncJT_OpenReader(strdes string, iComID int, baud int) int {
	cstr := C.CString(strdes)
	defer C.free(unsafe.Pointer(cstr))

	rlt := C.JT_OpenReader(cstr, (C.int)(iComID), (C.int)(baud))
	return int(rlt)
}

func (p *ReaderInf) FuncJT_CloseReader() int {
	rlt := C.JT_CloseReader()
	return int(rlt)
}

func (p *ReaderInf) FuncJT_GetLastError() (int, string) {
	cstr := C.CString("")
	defer C.free(unsafe.Pointer(cstr))

	rlt := C.JT_GetLastError(cstr)
	return int(rlt), C.GoString(cstr)
}

/*
func (p *ReaderInf) FuncJT_sErrorInfo(int ErrorCode) string {
	cstr := C.CString("")
	defer C.free(unsafe.Pointer(cstr))

	C.JT_sErrorInfo(ErrorCode, cstr)
	return C.GoString(cstr)
}
*/

func (p *ReaderInf) FuncJT_OpenCard() int {
	rlt := C.JT_OpenCard()
	return int(rlt)
}

func (p *ReaderInf) FuncJT_GetCardType() int {
	rlt := C.JT_GetCardType()
	return int(rlt)
}

func (p *ReaderInf) FuncJT_CloseCard() int {
	rlt := C.JT_CloseCard()
	return int(rlt)
}

func (p *ReaderInf) FuncJT_SelectPSAMSlot(iSlot int) int {
	rlt := C.JT_SelectPSAMSlot((C.int)(iSlot))
	return int(rlt)
}

func (p *ReaderInf) FuncJT_ProReadFile(FileID uint16, FileLen int) (int, []byte) {
	buf := make([]byte, 43)
	rlt := C.JT_ProReadFile((C.ushort)(FileID), (C.int)(FileLen), (*C.uchar)(&buf[0]))
	return int(rlt), buf
}

func (p *ReaderInf) FuncJT_ProWriteFile(FileID uint16, FileLen int, buf []byte) int {
	rlt := C.JT_ProWriteFile((C.ushort)(FileID), (C.int)(FileLen), (*C.uchar)(&buf[0]))
	return int(rlt)
}

func (p *ReaderInf) FuncJT_ProQueryBalance() (int, uint) {
	var balance C.uint
	rlt := C.JT_ProQueryBalance(&balance)
	return int(rlt), uint(balance)
}

func (p *ReaderInf) FuncJT_ProGetCardID() (int, []byte) {
	buf := make([]byte, 8)
	rlt := C.JT_ProGetCardID((*C.uchar)(&buf[0]))

	return int(rlt), buf
}

func (p *ReaderInf) FuncJT_ProGetCardType() int {
	rlt := C.JT_ProGetCardType()
	return int(rlt)
}

func (p *ReaderInf) FuncJT_ProDecrement(iMoney int, sTollInfo []byte, paytime string) (int, []byte, []byte, string, []byte) {
	cDectime := C.CString(paytime)
	defer C.free(unsafe.Pointer(cDectime))

	sTradNo := make([]byte, 2)
	sTermTradNo := make([]byte, 4)
	sTac := make([]byte, 4)

	rlt := C.JT_ProDecrement(C.int(iMoney), (*C.uchar)(&sTollInfo[0]), (*C.uchar)(&sTradNo[0]), (*C.uchar)(&sTermTradNo[0]), cDectime, (*C.uchar)(&sTac[0]))
	if int(rlt) == 0 {
		//支付成功，蜂鸣器响
		p.FuncJT_AudioControl(2, 3)
	}
	strDectime := C.GoString(cDectime)
	return int(rlt), sTradNo, sTermTradNo, strDectime, sTac
}

func (p *ReaderInf) FuncJT_SamGetSerialNo() (int, []byte) {
	buf := make([]byte, 10)
	rlt := C.JT_SamGetSerialNo((*C.uchar)(&buf[0]))

	return int(rlt), buf
}

func (p *ReaderInf) FuncJT_SamGetTermID() (int, []byte) {
	buf := make([]byte, 6)
	rlt := C.JT_SamGetTermID((*C.uchar)(&buf[0]))

	return int(rlt), buf
}

func (p *ReaderInf) FuncJT_SamReset() (int, int, []byte) {
	buf := make([]byte, 100)
	var buflen C.int

	rlt := C.JT_SamReset((*C.uchar)(&buf[0]), &buflen)

	return int(rlt), int(buflen), buf
}

func (p *ReaderInf) FuncJT_AudioControl(cTimes, cVoice byte) int {
	rlt := C.JT_AudioControl((C.uchar)(cTimes), (C.uchar)(cVoice))
	return int(rlt)
}

func (p *ReaderInf) FuncJT_ReaderVersion() (int, string) {
	cstr := C.CString("")
	defer C.free(unsafe.Pointer(cstr))

	rlt := C.JT_ReaderVersion(cstr)
	str := C.GoString(cstr)

	return int(rlt), str
}

func (p *ReaderInf) FuncJT_GetCardSer() []byte {
	buf := make([]byte, 4)

	C.JT_GetCardSer((*C.uchar)(&buf[0]))
	return buf
}

func (p *ReaderInf) FuncJT_ReadBlock(iKeyType, iBlockn int) (int, []byte) {
	buf := make([]byte, 16)

	rlt := C.JT_ReadBlock((C.int)(iKeyType), (C.int)(iBlockn), (*C.uchar)(&buf[0]))
	return int(rlt), buf
}

func (p *ReaderInf) FuncJT_WriteBlock(iKeyType, iBlockn int, sData []byte) int {
	rlt := C.JT_WriteBlock((C.int)(iKeyType), (C.int)(iBlockn), (*C.uchar)(&sData[0]))

	return int(rlt)
}

func (p *ReaderInf) FuncJT_GetCardID() (int, string) {
	cstr := C.CString("")
	defer C.free(unsafe.Pointer(cstr))

	rlt := C.JT_GetCardID(cstr)
	str := C.GoString(cstr)

	return int(rlt), str
}

func (p *ReaderInf) FuncJT_ReadFile(sFileID uint16, sKeyID byte, cFileType byte, iAddr int, iLength int) (int, []byte) {
	buf := make([]byte, iLength)
	rlt := C.JT_ReadFile((C.ushort)(sFileID), (C.uchar)(sKeyID), (C.uchar)(cFileType), (C.int)(iAddr), (C.int)(iLength), (*C.uchar)(&buf[0]))

	return int(rlt), buf
}

func (p *ReaderInf) FuncJT_WriteFile(sFileID uint16, sKeyID byte, cFileType byte, iAddr, iLength int, data []byte) int {
	rlt := C.JT_WriteFile((C.ushort)(sFileID), (C.uchar)(sKeyID), (C.uchar)(cFileType), (C.int)(iAddr), (C.int)(iLength), (*C.uchar)(&data[0]))

	return int(rlt)
}
