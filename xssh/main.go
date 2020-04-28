package main

import (
	"C"
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"sync"
	"time"
)

type SSHParam struct {
	User     string   `json:"user"`
	Password string   `json:"password"`
	Host     string   `json:"host"`
	Key      string   `json:"key"`
	Cmds     []string `json:"cmds"`
	Port     int      `json:"port"`
}

func connect(user, password, host, key string, port int) (*ssh.Session, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		config       ssh.Config
		session      *ssh.Session
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	if key == "" {
		auth = append(auth, ssh.Password(password))
	} else {
		if err != nil {
			return nil, err
		}

		var signer ssh.Signer
		if password == "" {
			signer, err = ssh.ParsePrivateKey([]byte(key))
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(key), []byte(password))
		}
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	config = ssh.Config{
		Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
	}

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		Config:  config,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create session
	if session, err = client.NewSession(); err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("term", 80, 40, modes); err != nil {
		return nil, err
	}

	return session, nil
}

func SshSession(s SSHParam, ch chan string, sync *sync.WaitGroup) {

	session, err := connect(s.User, s.Password, s.Host, s.Key, s.Port)

	if err != nil {
		ch <- fmt.Sprintf("connect host %s:<%s>", s.Host, err.Error())

		sync.Done()
		return
	}

	defer session.Close()

	cmdlist := append(s.Cmds, "exit")

	stdinBuf, _ := session.StdinPipe()

	var outbt, errbt bytes.Buffer
	session.Stdout = &outbt
	session.Stderr = &errbt

	err = session.Shell()
	if err != nil {
		ch <- fmt.Sprintf("host %s ,get shell session err: <%s>", s.Host, err.Error())

		sync.Done()
		return
	}

	for _, c := range cmdlist {
		c = c + "\n"
		stdinBuf.Write([]byte(c))
	}
	session.Wait()

	ch <- outbt.String()+"\n"+errbt.String()

	sync.Done()
	return
}

//export SSH
func SSH(sc *C.char) *C.char {

	// c 语言的char 转string
	sb := C.GoString(sc)

	var s []SSHParam

	if err := json.Unmarshal([]byte(sb), &s); err != nil {
		panic("json unmarshal err :" + err.Error())
	}

	var sync sync.WaitGroup

	c := make(chan string, len(s)+1)

	for _, v := range s {
		sync.Add(1)
		go SshSession(v, c, &sync)
	}
	//等待所有的 goroutine 执行完
	sync.Wait()

	close(c)

	var result []string
	for r := range c {
		result = append(result,r)
	}

	b,err:=json.Marshal(result)
	if err!=nil{
		panic("result json marshal err :" + err.Error())
	}

	return C.CString(string(b))
}

func main() {}
