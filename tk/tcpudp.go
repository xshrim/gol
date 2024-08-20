package tk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

// ICMP 数据包结构体
type ICMP struct {
	Type     uint8
	Code     uint8
	Checksum uint16
	ID       uint16
	Seq      uint16
}

// CheckSum 校验和计算
func CheckSum(data []byte) uint16 {
	var (
		sum    uint32
		length = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += (sum >> 16)
	return uint16(^sum)
}

func Ping(target string) (string, error) {
	raddr, _ := net.ResolveIPAddr("ip", target)

	//构建发送的ICMP包
	icmp := ICMP{
		Type:     8,
		Code:     0,
		Checksum: 0, //默认校验和为0，后面计算再写入
		ID:       0,
		Seq:      0,
	}

	//新建buffer将包内数据写入，以计算校验和并将校验和并存入icmp结构体中
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.Checksum = CheckSum(buffer.Bytes())
	buffer.Reset()
	//与目的ip地址建立连接，第二个参数为空则默认为本地ip，第三个参数为目的ip
	con, err := net.DialIP("ip4:icmp", nil, raddr)
	if err != nil {
		return "", err
	}
	//主函数接术后关闭连接
	defer con.Close()
	//构建buffer将要发送的数据存入
	var sendBuffer bytes.Buffer
	binary.Write(&sendBuffer, binary.BigEndian, icmp)
	if _, err := con.Write(sendBuffer.Bytes()); err != nil {
		return "", err
	}
	//开始计算时间
	timeStart := time.Now()
	//设置读取超时时间为2s
	con.SetReadDeadline((time.Now().Add(time.Second * 2)))
	//构建接受的比特数组
	rec := make([]byte, 1024)
	//读取连接返回的数据，将数据放入rec中
	recCnt, err := con.Read(rec)
	if err != nil {
		return "", err
	}
	//设置结束时间，计算两次时间之差为ping的时间
	timeEnd := time.Now()
	durationTime := timeEnd.Sub(timeStart).Nanoseconds() / 1e6
	//显示结果
	return fmt.Sprintf("%d bytes from %s: seq=%d time=%dms", recCnt, raddr.String(), icmp.Seq, durationTime), nil
}

// ping ip or domain
// func Ping(ip string) bool {
// 	type ICMP struct {
// 		Type        uint8
// 		Code        uint8
// 		Checksum    uint16
// 		Identifier  uint16
// 		SequenceNum uint16
// 	}

// 	icmp := ICMP{
// 		Type: 8,
// 	}

// 	recvBuf := make([]byte, 32)
// 	var buffer bytes.Buffer

// 	_ = binary.Write(&buffer, binary.BigEndian, icmp)
// 	icmp.Checksum = checkSum(buffer.Bytes())
// 	buffer.Reset()
// 	_ = binary.Write(&buffer, binary.BigEndian, icmp)

// 	Time, _ := time.ParseDuration("2s")
// 	conn, err := net.DialTimeout("ip4:icmp", ip, Time)
// 	if err != nil {
// 		return exec.Command("ping", ip, "-c", "2", "-i", "1", "-W", "3").Run() == nil
// 	}
// 	_, err = conn.Write(buffer.Bytes())
// 	if err != nil {
// 		return false
// 	}
// 	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 2))
// 	num, err := conn.Read(recvBuf)
// 	if err != nil {
// 		return false
// 	}

// 	_ = conn.SetReadDeadline(time.Time{})

// 	return string(recvBuf[0:num]) != ""
// }

// get local ipv4 address
func IPv4() []string {
	out := []string{"127.0.0.1"}
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				out = append(out, ipnet.IP.String())
			}
		}
	}
	return out
}

// get all ipv4 addresses in the range of the cidr
func Hosts(cidr string) []string {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); {
		ips = append(ips, ip.String())

		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}

	if len(ips) < 2 {
		return []string{}
	}
	return ips[1 : len(ips)-1]
}

func PortScan(target string, port int) (string, error) {
	raddr, _ := net.ResolveIPAddr("ip", target)
	addr := fmt.Sprintf("%s:%d", raddr, port)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		//fmt.Println(address, "是关闭的", err)
		return addr, err
	}
	defer conn.Close()
	return addr, nil
}

// Handles TC connection and perform synchorinization:
// TCP -> Stdout and Stdin -> TCP
func TcpHandle(con net.Conn) {
	stdout := streamCopy(con, os.Stdout)
	remote := streamCopy(os.Stdin, con)
	select {
	case <-stdout:
		fmt.Println("remote connection is closed")
	case <-remote:
		fmt.Println("local program is terminated")
	}
}

// Handle UDP connection
func UdpHandle(con net.Conn) {
	in := acceptFromUdpToStream(con, os.Stdout)
	fmt.Println("waiting for remote connection")
	remoteAddr := <-in
	fmt.Println("connected from", remoteAddr)
	out := putFromStreamToUdp(os.Stdin, con, remoteAddr)
	select {
	case <-in:
		fmt.Println("remote connection is closed")
	case <-out:
		fmt.Println("local program is terminated")
	}
}

// Performs copy operation between streams: os and tcp streams
func streamCopy(src io.Reader, dst io.Writer) <-chan int {
	buf := make([]byte, 1024)
	sync := make(chan int)
	go func() {
		defer func() {
			if con, ok := dst.(net.Conn); ok {
				con.Close()
				fmt.Printf("connection from %v is closed\n", con.RemoteAddr())
			}
			sync <- 0 // Notify that processing is finished
		}()
		for {
			var nBytes int
			var err error
			nBytes, err = src.Read(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Printf("read error: %s\n", err)
				}
				break
			}
			_, err = dst.Write(buf[0:nBytes])
			if err != nil {
				panic(fmt.Sprintf("write error: %s\n", err))
			}
		}
	}()
	return sync
}

// Accept data from UPD connection and copy it to the stream
func acceptFromUdpToStream(src net.Conn, dst io.Writer) <-chan net.Addr {
	buf := make([]byte, 1024)
	sync := make(chan net.Addr)
	con, ok := src.(*net.UDPConn)
	if !ok {
		fmt.Printf("input must be UDP connection")
		return sync
	}
	go func() {
		var remoteAddr net.Addr
		for {
			var nBytes int
			var err error
			var addr net.Addr
			nBytes, addr, err = con.ReadFromUDP(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Printf("read error: %s\n", err)
				}
				break
			}
			if remoteAddr == nil && remoteAddr != addr {
				remoteAddr = addr
				sync <- remoteAddr
			}
			_, err = dst.Write(buf[0:nBytes])
			if err != nil {
				panic(fmt.Sprintf("write error: %s\n", err))
			}
		}
	}()
	fmt.Println("exit write from udp to stream")
	return sync
}

// Put input date from the stream to UDP connection
func putFromStreamToUdp(src io.Reader, dst net.Conn, remoteAddr net.Addr) <-chan net.Addr {
	buf := make([]byte, 1024)
	sync := make(chan net.Addr)
	go func() {
		for {
			var nBytes int
			var err error
			nBytes, err = src.Read(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Printf("read error: %s\n", err)
				}
				break
			}
			fmt.Println("write to the remote address:", remoteAddr)
			if con, ok := dst.(*net.UDPConn); ok && remoteAddr != nil {
				_, err = con.WriteTo(buf[0:nBytes], remoteAddr)
			}
			if err != nil {
				panic(fmt.Sprintf("write error: %s\n", err))
			}
		}
	}()
	return sync
}
