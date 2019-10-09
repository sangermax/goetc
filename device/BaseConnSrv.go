package device

import (
	"FTC/models"
	"FTC/util"
	"net"
)

type ConnBaseSrv struct {
	ConnType int
	ConnUrl  string
}

func (p *ConnBaseSrv) InitConn(itype int, strconn string) {
	p.ConnType = itype
	p.ConnUrl = strconn

	go p.InitConnSrv()
}

func (p *ConnBaseSrv) InitConnSrv() {
	listen_sock, err := net.Listen("tcp", p.ConnUrl)
	if err != nil {
		util.FileLogs.Info("%d InitConnSrv listen failed.", p.ConnType)
		return
	}
	defer listen_sock.Close()

	for {
		new_conn, err := listen_sock.Accept()
		if err != nil {
			continue
		}

		go p.goConnCliProc(new_conn)
	}
}

func (p *ConnBaseSrv) SendProc(conn net.Conn, sendbuf []byte) bool {
	_, err := conn.Write(sendbuf)
	if err != nil {
		util.FileLogs.Info("%d,ConnBaseSrv SendProc failed:%s.\r\n", p.ConnType, err.Error())
		return false
	}

	//util.FileLogs.Info("%d,ConnBaseSrv SendProc suc:%d,%d.\r\n",p.ConnType,n,len(sendbuf))
	return true
}

func (p *ConnBaseSrv) goConnCliProc(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, models.MAX_BUFFERSIZE)
	buflen := 0

	for {
		if buflen < 0 || buflen > models.MAX_BUFFERSIZE {
			util.FileLogs.Info("%d,ConnBaseSrv run:%d.\r\n", p.ConnType, buflen)
			buflen = 0
		}

		num, err := conn.Read(buf[buflen:])
		if err != nil {
			util.FileLogs.Info("%d,ConnBaseSrv Read failed:%s.\r\n", p.ConnType, err.Error())
			return
		}
		buflen += num

		for {
			if buflen <= 0 {
				break
			}

			sublen := 0
			switch p.ConnType {

			default:
				sublen = p.Recvproc(buf[0:buflen])
			}

			copy(buf, buf[sublen:])
			buflen -= sublen

		}

		util.MySleep_ms(10)
	}
}

func (p *ConnBaseSrv) Recvproc(inbuf []byte) int {
	inlen := len(inbuf)

	util.FileLogs.Info("ConnBaseSrv:%d,%s.\r\n", inlen, util.ConvertByte2Hexstring(inbuf, false))

	return inlen
}
