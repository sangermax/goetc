package device

import (
	"FTC/models"
	"FTC/util"
	"net"
	"time"
)

type ConnRecvProc func([]byte) (int, bool)

type ConnBaseCli struct {
	ConnType   int
	ConnUrl    string
	ConnHandle net.Conn
	Connstate  bool
}

func (p *ConnBaseCli) InitConn(itype int, strconn string, funcRecv ConnRecvProc) {
	p.ConnType = itype
	p.ConnUrl = strconn
	p.ConnHandle = nil
	p.Connstate = false

	go p.GoRun(funcRecv)
}

func (p *ConnBaseCli) IsConn() bool {
	return p.Connstate
}

//客户端连接
func (p *ConnBaseCli) ConnectTo() (bool, error) {
	p.DisConnect()

	var err error
	p.ConnHandle, err = net.Dial("tcp", p.ConnUrl)
	if err != nil {
		util.FileLogs.Info("(%d-%s) 连接外设服务失败", p.ConnType, GetDevSrvDes(p.ConnType))
		return false, err
	}

	util.FileLogs.Info("(%d-%s)  连接外设服务成功", p.ConnType, GetDevSrvDes(p.ConnType))
	p.Connstate = true
	return true, nil
}

func (p *ConnBaseCli) DisConnect() {
	if p.ConnHandle != nil {
		p.ConnHandle.Close()

		util.FileLogs.Info("(%d-%s)  断开外设服务", p.ConnType, GetDevSrvDes(p.ConnType))
	}

	p.Connstate = false
}

func (p *ConnBaseCli) SendProc(sendbuf []byte, bshow bool) bool {
	if !p.Connstate || sendbuf == nil {
		util.FileLogs.Info("(%d-%s) ,ConnBaseCli SendProc failed:nil.\r\n", p.ConnType, GetDevSrvDes(p.ConnType))
		return false
	}

	n, err := p.ConnHandle.Write(sendbuf)
	if err != nil {
		util.FileLogs.Info("(%d-%s) ,ConnBaseCli SendProc failed:%s.\r\n", p.ConnType, GetDevSrvDes(p.ConnType), err.Error())
		return false
	}

	if bshow {
		util.FileLogs.Info("%d,ConnBaseCli SendProc suc:%d,%d.\r\n", p.ConnType, n, len(sendbuf))
	}

	return true
}

func (p *ConnBaseCli) GoRun(funcRecv ConnRecvProc) {
	defer p.DisConnect()

	buf := make([]byte, models.MAX_BUFFERSIZE)
	buflen := 0

	for {
		if !p.Connstate {
			p.ConnectTo()
			time.Sleep(5 * time.Second)
			continue
		}

		if buflen < 0 || buflen > models.MAX_BUFFERSIZE {
			util.FileLogs.Info("(%d-%s) ,ConnBaseCli run:%d.\r\n", p.ConnType, GetDevSrvDes(p.ConnType), buflen)
			buflen = 0
		}

		num, err := p.ConnHandle.Read(buf[buflen:])
		if err != nil {
			util.FileLogs.Info("(%d-%s) ,ConnBaseCli Read failed:%s.\r\n", p.ConnType, GetDevSrvDes(p.ConnType), err.Error())
			p.DisConnect()
			util.MySleep_ms(10)
			continue
		}

		buflen += num

		for {
			if buflen <= 0 {
				break
			}

			sublen, bflag := funcRecv(buf[0:buflen])
			copy(buf, buf[sublen:])
			buflen -= sublen

			if !bflag {
				break
			}

		}

		util.MySleep_ms(10)
	}
}

func (p *ConnBaseCli) Recvproc(inbuf []byte) int {
	inlen := len(inbuf)

	util.FileLogs.Info("ConnBaseCli:%d,%s.\r\n", inlen, util.ConvertByte2Hexstring(inbuf, false))

	return inlen
}
