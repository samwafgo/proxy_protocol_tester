package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
)

// 简单的测试服务器，用于接收和解析 Proxy Protocol
func runTestServer() {
	if len(os.Args) < 3 || os.Args[1] != "server" {
		return
	}

	port := "8080"
	if len(os.Args) > 2 {
		port = os.Args[2]
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("启动服务器失败: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("测试服务器启动在端口 %s\n", port)
	fmt.Println("等待连接...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("接受连接失败: %v\n", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Printf("新连接来自: %s\n", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	// 尝试解析 Proxy Protocol
	if err := parseProxyProtocol(reader); err != nil {
		fmt.Printf("解析 Proxy Protocol 失败: %v\n", err)
		return
	}

	// 读取剩余数据
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Printf("收到数据: %s", line)
		if strings.TrimSpace(line) == "" {
			break
		}
	}

	// 发送简单响应
	response := "HTTP/1.1 200 OK\r\nContent-Length: 13\r\n\r\nHello, World!"
	conn.Write([]byte(response))

	fmt.Println("连接处理完成")
}

func parseProxyProtocol(reader *bufio.Reader) error {
	// 先读取一些字节来判断版本
	peek, err := reader.Peek(16)
	if err != nil {
		return fmt.Errorf("读取数据失败: %v", err)
	}

	if strings.HasPrefix(string(peek), "PROXY") {
		return parseProxyV1(reader)
	} else if len(peek) >= 12 && string(peek[:12]) == ProxyV2Signature {
		return parseProxyV2(reader)
	}

	return fmt.Errorf("未检测到 Proxy Protocol 头部")
}

func parseProxyV1(reader *bufio.Reader) error {
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("读取 Proxy Protocol v1 头部失败: %v", err)
	}

	line = strings.TrimSpace(line)
	fmt.Printf("Proxy Protocol v1 头部: %s\n", line)

	parts := strings.Split(line, " ")
	if len(parts) != 6 {
		return fmt.Errorf("Proxy Protocol v1 格式错误")
	}

	fmt.Printf("协议: %s\n", parts[1])
	fmt.Printf("源IP: %s\n", parts[2])
	fmt.Printf("目标IP: %s\n", parts[3])
	fmt.Printf("源端口: %s\n", parts[4])
	fmt.Printf("目标端口: %s\n", parts[5])

	return nil
}

func parseProxyV2(reader *bufio.Reader) error {
	// 读取固定头部 (16 bytes)
	header := make([]byte, 16)
	if _, err := reader.Read(header); err != nil {
		return fmt.Errorf("读取 Proxy Protocol v2 头部失败: %v", err)
	}

	fmt.Printf("Proxy Protocol v2 头部 (hex): %x\n", header)

	// 验证签名
	if string(header[:12]) != ProxyV2Signature {
		return fmt.Errorf("Proxy Protocol v2 签名错误")
	}

	versionCmd := header[12]
	famTransport := header[13]
	length := binary.BigEndian.Uint16(header[14:16])

	fmt.Printf("版本和命令: 0x%02x\n", versionCmd)
	fmt.Printf("地址族和传输协议: 0x%02x\n", famTransport)
	fmt.Printf("地址信息长度: %d\n", length)

	// 读取地址信息
	if length > 0 {
		addrInfo := make([]byte, length)
		if _, err := reader.Read(addrInfo); err != nil {
			return fmt.Errorf("读取地址信息失败: %v", err)
		}

		if err := parseAddressInfo(famTransport, addrInfo); err != nil {
			return fmt.Errorf("解析地址信息失败: %v", err)
		}
	}

	return nil
}

func parseAddressInfo(famTransport byte, addrInfo []byte) error {
	family := famTransport & 0xF0
	transport := famTransport & 0x0F

	var transportStr string
	switch transport {
	case ProxyV2TransStream:
		transportStr = "TCP"
	case ProxyV2TransDgram:
		transportStr = "UDP"
	default:
		transportStr = "Unknown"
	}

	fmt.Printf("传输协议: %s\n", transportStr)

	switch family {
	case ProxyV2FamInet:
		if len(addrInfo) < 12 {
			return fmt.Errorf("IPv4 地址信息长度不足")
		}
		srcIP := net.IP(addrInfo[0:4])
		dstIP := net.IP(addrInfo[4:8])
		srcPort := binary.BigEndian.Uint16(addrInfo[8:10])
		dstPort := binary.BigEndian.Uint16(addrInfo[10:12])

		fmt.Printf("地址族: IPv4\n")
		fmt.Printf("源地址: %s:%d\n", srcIP, srcPort)
		fmt.Printf("目标地址: %s:%d\n", dstIP, dstPort)

	case ProxyV2FamInet6:
		if len(addrInfo) < 36 {
			return fmt.Errorf("IPv6 地址信息长度不足")
		}
		srcIP := net.IP(addrInfo[0:16])
		dstIP := net.IP(addrInfo[16:32])
		srcPort := binary.BigEndian.Uint16(addrInfo[32:34])
		dstPort := binary.BigEndian.Uint16(addrInfo[34:36])

		fmt.Printf("地址族: IPv6\n")
		fmt.Printf("源地址: [%s]:%d\n", srcIP, srcPort)
		fmt.Printf("目标地址: [%s]:%d\n", dstIP, dstPort)

	default:
		fmt.Printf("地址族: Unknown (0x%02x)\n", family)
	}

	return nil
}

func init() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		runTestServer()
		os.Exit(0)
	}
}
