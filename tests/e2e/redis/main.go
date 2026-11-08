package main
import("fmt";"os";"os/signal";"github.com/alicebob/miniredis/v2")
func main(){server,err:=miniredis.Run();if err!=nil{panic(err)};defer server.Close();fmt.Println(server.Addr());done:=make(chan os.Signal,1);signal.Notify(done,os.Interrupt);<-done}
