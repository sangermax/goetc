package device

import (
	"FTC/models"
	"FTC/util"
	"io"
	"time"

	serial "github.com/tarm/goserial"
)

//收到应答处理
type SerialRecvProc func([]byte) (int, error)

//连接初始化
type SerialConnectedInit func()

//串口基类
type SerialBase struct {
	devType   int
	ComDes    string
	ComBaud   int
	ComHandle io.ReadWriteCloser //串口句柄
	ComState  bool
}

func (p *SerialBase) InitSerial(itype int, strcom string, ibaud int, funcrecv SerialRecvProc) bool {
	p.devType = itype
	p.ComDes = strcom
	p.ComBaud = ibaud
	p.ComHandle = nil
	p.ComState = false

	if p.OpenSerial() {
		go p.GoRun(funcrecv)
		return true
	}

	return false
}

func (p *SerialBase) InitSerial2(itype int, strcom string, ibaud int, funcInit SerialConnectedInit, funcrecv SerialRecvProc) bool {
	p.devType = itype
	p.ComDes = strcom
	p.ComBaud = ibaud
	p.ComHandle = nil
	p.ComState = false

	if p.OpenSerial() {
		funcInit()
		go p.GoRun(funcrecv)
		return true
	}

	return false
}

func (p *SerialBase) OpenSerial() bool {
	p.CloseSerial()

	var err error
	c := &serial.Config{Name: p.ComDes, Baud: p.ComBaud}
	if c == nil {
		util.FileLogs.Info("(%d-%s) OpenSerial set serial params failed", p.devType, GetDevSrvDes(p.devType))
		return false
	}

	//打开串口
	p.ComHandle, err = serial.OpenPort(c)
	if err != nil {
		util.FileLogs.Info("(%d-%s) OpenSerial failed.(%s-%d),err:%s.\r\n", p.devType, GetDevSrvDes(p.devType), p.ComDes, p.ComBaud, err.Error())
		return false
	}

	util.FileLogs.Info("(%d-%s) OpenSerial suc.(%s-%d) suc.\r\n", p.devType, GetDevSrvDes(p.devType), p.ComDes, p.ComBaud)
	p.ComState = true

	return true
}

func (p *SerialBase) CloseSerial() {
	if p.ComHandle != nil {
		p.ComHandle.Close()
		p.ComHandle = nil
	}

	p.ComState = false
}

func (p *SerialBase) IsConn() bool {
	return p.ComState
}

func (p *SerialBase) SendProc(sendbuf []byte, bshow bool) bool {
	if !p.ComState || sendbuf == nil {
		util.FileLogs.Info("(%d-%s) ,SerialBase SendProc failed:nil or closed.\r\n", p.devType, GetDevSrvDes(p.devType))
		return false
	}

	n, err := p.ComHandle.Write(sendbuf)
	if err != nil {
		util.FileLogs.Info("(%d-%s) ,SerialBase SendProc failed:%s.\r\n", p.devType, GetDevSrvDes(p.devType), err.Error())
		return false
	}

	if bshow {
		util.FileLogs.Info("(%d-%s) ,SerialBase SendProc suc:%d,%d,%s.\r\n", p.devType, GetDevSrvDes(p.devType), n, len(sendbuf), util.ConvertByte2Hexstring(sendbuf, true))
	}

	return true
}

func (p *SerialBase) GoRun(funcRecv SerialRecvProc) {
	defer p.CloseSerial()

	buf := make([]byte, models.MAX_BUFFERSIZE)
	buflen := 0

	for {
		if p.ComHandle == nil {
			return
		}

		if buflen < 0 || buflen > models.MAX_BUFFERSIZE {
			util.FileLogs.Info("(%d-%s) ,SerialBase run:%d.\r\n", p.devType, GetDevSrvDes(p.devType), buflen)
			buflen = 0
		}

		num, err := p.ComHandle.Read(buf[buflen:])
		if err != nil {
			util.FileLogs.Info("(%d-%s) ,SerialBase run read failed:%s.\r\n", p.devType, GetDevSrvDes(p.devType), err.Error())

			time.Sleep(10 * time.Second)
			continue
		}

		buflen += num

		for {
			if buflen <= 0 {
				break
			}

			sublen, err := funcRecv(buf[0:buflen])
			copy(buf, buf[sublen:])
			buflen -= sublen

			if err != nil {
				break
			}
		}

		util.MySleep_ms(10)
	}
}
