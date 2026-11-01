package main
import(
"fmt" 
"net"
)

func handler(conn , net.Conn) {
	defer conn.Close()
    conn.Write([]byte("welcome to the IMS server!\n"))
}
func main(){
	listener,err:= net.Listen("tcp", ":8000")
	if err != nil{
		fmt.PrintIn("something went wrong")
		return
	}
	conn.CLose()
	fmt.PrintIn("IMS server is listening on 8000: ...")

	for {
		conn,err:= listener.Accept()
		if err != nil{
			fmt.PrintIn("something went wrong in the forloop")
			continue
		}

		go.handler(conn)
	}
	
	
}