package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// Proxy Protocol v1 signature
	ProxyV1Signature = "PROXY"

	// Proxy Protocol v2 signature
	ProxyV2Signature = "\x0D\x0A\x0D\x0A\x00\x0D\x0A\x51\x55\x49\x54\x0A"

	// Protocol versions
	ProxyV2VersionCmd = 0x21
	ProxyV2Local      = 0x20

	// Address families
	ProxyV2FamUnspec = 0x00
	ProxyV2FamInet   = 0x10
	ProxyV2FamInet6  = 0x20
	ProxyV2FamUnix   = 0x30

	// Transport protocols
	ProxyV2TransUnspec = 0x00
	ProxyV2TransStream = 0x01
	ProxyV2TransDgram  = 0x02
)

type Config struct {
	Version    int
	ServerAddr string
	ServerPort int
	SrcIP      string
	SrcPort    int
	DstIP      string
	DstPort    int
	Protocol   string
	Message    string
	Timeout    int
}

func main() {

	// 如果有命令行参数，使用命令行模式
	if len(os.Args) > 1 {
		config := parseFlags()
		runTest(config)
		return
	}

	// 交互式模式
	config := interactiveConfig()
	runTest(config)
}

func interactiveConfig() *Config {
	config := &Config{}
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Proxy Protocol 测试工具 ===")
	fmt.Println("欢迎使用交互式配置模式！")
	fmt.Println()

	// 选择版本
	for {
		fmt.Print("请选择 Proxy Protocol 版本 (1 或 2) [默认: 1]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			config.Version = 1
			break
		}

		if version, err := strconv.Atoi(input); err == nil && (version == 1 || version == 2) {
			config.Version = version
			break
		}

		fmt.Println("❌ 请输入 1 或 2")
	}

	// 服务器地址
	fmt.Print("请输入目标服务器地址 [默认: 127.0.0.1]: ")
	input, _ := reader.ReadString('\n')
	config.ServerAddr = strings.TrimSpace(input)
	if config.ServerAddr == "" {
		config.ServerAddr = "127.0.0.1"
	}

	// 服务器端口
	for {
		fmt.Print("请输入目标服务器端口 [默认: 8080]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			config.ServerPort = 8080
			break
		}

		if port, err := strconv.Atoi(input); err == nil && port > 0 && port <= 65535 {
			config.ServerPort = port
			break
		}

		fmt.Println("❌ 请输入有效的端口号 (1-65535)")
	}

	fmt.Println()
	fmt.Println("--- 配置源地址信息 (模拟的真实客户端) ---")

	// 源IP
	fmt.Print("请输入源IP地址 [默认: 192.168.1.100]: ")
	input, _ = reader.ReadString('\n')
	config.SrcIP = strings.TrimSpace(input)
	if config.SrcIP == "" {
		config.SrcIP = "192.168.1.100"
	}

	// 源端口
	for {
		fmt.Print("请输入源端口 [默认: 12345]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			config.SrcPort = 12345
			break
		}

		if port, err := strconv.Atoi(input); err == nil && port > 0 && port <= 65535 {
			config.SrcPort = port
			break
		}

		fmt.Println("❌ 请输入有效的端口号 (1-65535)")
	}

	fmt.Println()
	fmt.Println("--- 配置目标地址信息 ---")

	// 目标IP
	fmt.Print("请输入目标IP地址 [默认: 192.168.1.200]: ")
	input, _ = reader.ReadString('\n')
	config.DstIP = strings.TrimSpace(input)
	if config.DstIP == "" {
		config.DstIP = "192.168.1.200"
	}

	// 目标端口
	for {
		fmt.Print("请输入目标端口 [默认: 80]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			config.DstPort = 80
			break
		}

		if port, err := strconv.Atoi(input); err == nil && port > 0 && port <= 65535 {
			config.DstPort = port
			break
		}

		fmt.Println("❌ 请输入有效的端口号 (1-65535)")
	}

	// 协议类型
	for {
		fmt.Print("请选择协议类型 (TCP/UDP) [默认: TCP]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToUpper(input))

		if input == "" {
			config.Protocol = "TCP"
			break
		}

		if input == "TCP" || input == "UDP" {
			config.Protocol = input
			break
		}

		fmt.Println("❌ 请输入 TCP 或 UDP")
	}

	// 测试消息
	fmt.Println()
	fmt.Println("--- 配置测试消息 ---")
	fmt.Println("请选择测试消息类型:")
	fmt.Println("1. HTTP GET 请求 (默认)")
	fmt.Println("2. 自定义消息")
	fmt.Println("3. 不发送消息")

	for {
		fmt.Print("请选择 (1-3) [默认: 1]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "", "1":
			// 让用户输入域名
			fmt.Print("请输入 Host 域名 [默认: example.com]: ")
			hostInput, _ := reader.ReadString('\n')
			hostInput = strings.TrimSpace(hostInput)
			if hostInput == "" {
				hostInput = "example.com"
			}
			config.Message = fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", hostInput)
			goto messageConfigDone
		case "2":
			fmt.Print("请输入自定义消息: ")
			customMsg, _ := reader.ReadString('\n')
			config.Message = strings.TrimSpace(customMsg)
			goto messageConfigDone
		case "3":
			config.Message = ""
			goto messageConfigDone
		default:
			fmt.Println("❌ 请输入 1、2 或 3")
		}
	}

messageConfigDone:

	// 超时设置
	for {
		fmt.Print("请输入连接超时时间(秒) [默认: 10]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			config.Timeout = 10
			break
		}

		if timeout, err := strconv.Atoi(input); err == nil && timeout > 0 {
			config.Timeout = timeout
			break
		}

		fmt.Println("❌ 请输入有效的超时时间")
	}

	return config
}

func runTest(config *Config) {
	fmt.Println()
	fmt.Println("=== 测试配置确认 ===")
	fmt.Printf("Proxy Protocol 版本: v%d\n", config.Version)
	fmt.Printf("目标服务器: %s:%d\n", config.ServerAddr, config.ServerPort)
	fmt.Printf("源地址: %s:%d\n", config.SrcIP, config.SrcPort)
	fmt.Printf("目标地址: %s:%d\n", config.DstIP, config.DstPort)
	fmt.Printf("协议: %s\n", config.Protocol)
	if config.Message != "" {
		fmt.Printf("测试消息: %q\n", config.Message)
	} else {
		fmt.Println("测试消息: (无)")
	}
	fmt.Printf("超时时间: %d秒\n", config.Timeout)
	fmt.Println(strings.Repeat("=", 50))

	// 询问是否继续
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("确认开始测试? (y/N): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		fmt.Println("测试已取消")
		return
	}

	fmt.Println()
	fmt.Println("🚀 开始测试...")
	fmt.Println()

	if err := testProxyProtocol(config); err != nil {
		fmt.Printf("❌ 测试失败: %v\n", err)

		// 提供帮助信息
		fmt.Println()
		fmt.Println("💡 故障排除建议:")
		fmt.Println("1. 检查目标服务器是否启动并监听指定端口")
		fmt.Println("2. 确认服务器支持 Proxy Protocol")
		fmt.Println("3. 检查网络连接和防火墙设置")
		fmt.Println("4. 尝试使用内置测试服务器: go run . server")

		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ 测试完成!")

	// 询问是否再次测试
	fmt.Print("是否进行另一次测试? (y/N): ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		fmt.Println()
		newConfig := interactiveConfig()
		runTest(newConfig)
	}
}

func parseFlags() *Config {
	config := &Config{}

	flag.IntVar(&config.Version, "version", 1, "Proxy Protocol 版本 (1 或 2)")
	flag.StringVar(&config.ServerAddr, "server", "127.0.0.1", "目标服务器地址")
	flag.IntVar(&config.ServerPort, "port", 8080, "目标服务器端口")
	flag.StringVar(&config.SrcIP, "src-ip", "192.168.1.100", "源IP地址")
	flag.IntVar(&config.SrcPort, "src-port", 12345, "源端口")
	flag.StringVar(&config.DstIP, "dst-ip", "192.168.1.200", "目标IP地址")
	flag.IntVar(&config.DstPort, "dst-port", 80, "目标端口")
	flag.StringVar(&config.Protocol, "protocol", "TCP", "传输协议 (TCP/UDP)")
	flag.StringVar(&config.Message, "message", "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n", "发送的测试消息")
	flag.IntVar(&config.Timeout, "timeout", 10, "连接超时时间(秒)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Proxy Protocol v1/v2 测试工具\n\n")
		fmt.Fprintf(os.Stderr, "用法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  # 测试 Proxy Protocol v1\n")
		fmt.Fprintf(os.Stderr, "  %s -version=1 -server=192.168.1.10 -port=8080\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 测试 Proxy Protocol v2\n")
		fmt.Fprintf(os.Stderr, "  %s -version=2 -server=192.168.1.10 -port=8080 -src-ip=10.0.0.1 -dst-ip=10.0.0.2\n\n", os.Args[0])
	}

	flag.Parse()

	if config.Version != 1 && config.Version != 2 {
		fmt.Fprintf(os.Stderr, "错误: 版本必须是 1 或 2\n")
		os.Exit(1)
	}

	return config
}

func testProxyProtocol(config *Config) error {
	// 连接到目标服务器
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", config.ServerAddr, config.ServerPort),
		time.Duration(config.Timeout)*time.Second)
	if err != nil {
		return fmt.Errorf("连接服务器失败: %v", err)
	}
	defer conn.Close()

	fmt.Println("已连接到服务器")

	// 发送 Proxy Protocol 头部
	var header []byte
	if config.Version == 1 {
		header = buildProxyV1Header(config)
		fmt.Printf("发送 Proxy Protocol v1 头部: %q\n", string(header))
	} else {
		header = buildProxyV2Header(config)
		fmt.Printf("发送 Proxy Protocol v2 头部 (%d 字节)\n", len(header))
		fmt.Printf("头部内容 (hex): %x\n", header)
	}

	if _, err := conn.Write(header); err != nil {
		return fmt.Errorf("发送 Proxy Protocol 头部失败: %v", err)
	}

	fmt.Println("Proxy Protocol 头部发送成功")

	// 发送测试消息
	if config.Message != "" {
		fmt.Printf("发送测试消息: %q\n", config.Message)
		if _, err := conn.Write([]byte(config.Message)); err != nil {
			return fmt.Errorf("发送测试消息失败: %v", err)
		}
	}

	// 读取响应
	conn.SetReadDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
	reader := bufio.NewReader(conn)

	fmt.Println("等待服务器响应...")
	response := make([]byte, 4096)
	n, err := reader.Read(response)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	fmt.Printf("收到响应 (%d 字节):\n%s\n", n, string(response[:n]))

	return nil
}

func buildProxyV1Header(config *Config) []byte {
	// PROXY TCP4/TCP6 srcIP destIP srcPort destPort\r\n
	protocol := "TCP4"
	if strings.Contains(config.SrcIP, ":") {
		protocol = "TCP6"
	}
	if strings.ToUpper(config.Protocol) == "UDP" {
		if strings.Contains(config.SrcIP, ":") {
			protocol = "UDP6"
		} else {
			protocol = "UDP4"
		}
	}

	header := fmt.Sprintf("PROXY %s %s %s %d %d\r\n",
		protocol, config.SrcIP, config.DstIP, config.SrcPort, config.DstPort)

	return []byte(header)
}

func buildProxyV2Header(config *Config) []byte {
	// Proxy Protocol v2 binary format
	header := make([]byte, 0, 256)

	// Signature (12 bytes)
	header = append(header, []byte(ProxyV2Signature)...)

	// Version and Command (1 byte)
	header = append(header, ProxyV2VersionCmd)

	// Address family and transport protocol (1 byte)
	var famAndTransport byte
	if strings.Contains(config.SrcIP, ":") {
		famAndTransport = ProxyV2FamInet6
	} else {
		famAndTransport = ProxyV2FamInet
	}

	if strings.ToUpper(config.Protocol) == "UDP" {
		famAndTransport |= ProxyV2TransDgram
	} else {
		famAndTransport |= ProxyV2TransStream
	}

	header = append(header, famAndTransport)

	// Address information
	var addrInfo []byte
	if strings.Contains(config.SrcIP, ":") {
		// IPv6
		srcIP := net.ParseIP(config.SrcIP).To16()
		dstIP := net.ParseIP(config.DstIP).To16()
		addrInfo = append(addrInfo, srcIP...)
		addrInfo = append(addrInfo, dstIP...)

		// Ports (2 bytes each, big endian)
		srcPortBytes := make([]byte, 2)
		dstPortBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(srcPortBytes, uint16(config.SrcPort))
		binary.BigEndian.PutUint16(dstPortBytes, uint16(config.DstPort))
		addrInfo = append(addrInfo, srcPortBytes...)
		addrInfo = append(addrInfo, dstPortBytes...)
	} else {
		// IPv4
		srcIP := net.ParseIP(config.SrcIP).To4()
		dstIP := net.ParseIP(config.DstIP).To4()
		addrInfo = append(addrInfo, srcIP...)
		addrInfo = append(addrInfo, dstIP...)

		// Ports (2 bytes each, big endian)
		srcPortBytes := make([]byte, 2)
		dstPortBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(srcPortBytes, uint16(config.SrcPort))
		binary.BigEndian.PutUint16(dstPortBytes, uint16(config.DstPort))
		addrInfo = append(addrInfo, srcPortBytes...)
		addrInfo = append(addrInfo, dstPortBytes...)
	}

	// Length (2 bytes, big endian)
	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, uint16(len(addrInfo)))
	header = append(header, lengthBytes...)

	// Address information
	header = append(header, addrInfo...)

	return header
}
