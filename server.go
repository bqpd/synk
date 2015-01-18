/*

        GGGGGGGGGGGGG
     GGG::::::::::::G
   GG:::::::::::::::G
  G:::::GGGGGGGG::::G
 G:::::G       GGGGGG   ooooooooooo
G:::::G               oo:::::::::::oo
G:::::G              o:::::::::::::::o
G:::::G    GGGGGGGGGGo:::::ooooo:::::o
G:::::G    G::::::::Go::::o     o::::o
G:::::G    GGGGG::::Go::::o     o::::o
G:::::G        G::::Go::::o     o::::o
 G:::::G       G::::Go::::o     o::::o
  G:::::GGGGGGGG::::Go:::::ooooo:::::o
   GG:::::::::::::::Go:::::::::::::::o
     GGG::::::GGG:::G oo:::::::::::oo
        GGGGGG   GGGG   ooooooooooo

*/

package main

import (
    //"os"
    "fmt"
    "io"
    "log"
    "net/http"
    "code.google.com/p/go.net/websocket"
    "sync"
    "path/filepath"
    "strconv"
    "math/rand"
)

func main() {
    http.HandleFunc("/", rootHandler)
    http.Handle("/socket/", websocket.Handler(socketHandler))
    //err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
    err := http.ListenAndServe(":4000", nil)
    if err != nil {
        log.Fatal(err)
    }
}

type socket struct {
    io.ReadWriter
    ws *websocket.Conn
    done chan bool
    loc string
    id string
}

func (s socket) Close() error {
    s.done <- true
    return nil
}

var socketmap = make( map[string]chan socket )

var checkingSocketMap = new(sync.Mutex)

func socketHandler(ws *websocket.Conn) {
    loc := ws.Config().Location.String()
    var id string
    websocket.Message.Receive(ws, &id)
    s := socket{ws, ws, make(chan bool), loc, id}

    checkingSocketMap.Lock()
    if _, exist := socketmap[loc]; !exist {
        socketmap[loc] = make(chan socket)
    }
    checkingSocketMap.Unlock()

    go match(s)

    <-s.done
    fmt.Println("[ws] closing connection to "+id+" on channel "+loc)
}

func match(c socket) {
    fmt.Println("[ws] "+c.id+" added to channel "+c.loc)
    fmt.Fprint(c, "/sys Waiting for a partner...")
    select {
    case socketmap[c.loc] <- c:
        // now handled by the other goroutine
    case p := <-socketmap[c.loc]:
        if p.id != c.id {
            chat(p, c)
        } else {
            match(c)
        }
    }
}

func chat(a, b socket) {
    fmt.Println("[ws] matched "+a.id+" and "+b.id+" on channel "+a.loc)
    var x, y, z int

    //fmt.Fprint(a, "/sys You are talking to "+b.id)
    //fmt.Fprint(b, "/sys You are talking to "+a.id)

    var score int = 0

    for i := 0; i < 100; i++ {

        x = rand.Intn(99)
        y = rand.Intn(99)
        z = rand.Intn(99)
        fmt.Fprint(a, "INIT "+strconv.Itoa(x)+" "+strconv.Itoa(y)+" "+strconv.Itoa(z))
        fmt.Fprint(b, "INIT "+strconv.Itoa(x)+" "+strconv.Itoa(y)+" "+strconv.Itoa(z))

        var A, B string
        for A == "" || B == "" {
            var temp string
            websocket.Message.Receive(a.ws, &temp)
            if temp != ">heartbeat<" { A = temp }
            fmt.Println("A sent: " + temp + ", A is: " + A)

            websocket.Message.Receive(b.ws, &temp)
            if temp != ">heartbeat<" { B = temp }
            fmt.Println("B sent: " + temp + ", B is: " + B)
        }

        fmt.Println(strconv.Itoa(i) + " results: " + A + ", " + B)

        if A == B {
            fmt.Println("SUCCESS ON CHANNEL "+a.loc)
            fmt.Fprint(a, "GOOD")
            fmt.Fprint(b, "GOOD")
            score++
        } else {
            fmt.Println("FAILURE ON CHANNEL "+a.loc)
            fmt.Fprint(a, "FAIL")
            fmt.Fprint(b, "FAIL")
        }

        fmt.Println("SCORE FOR "+a.id+" AND "+b.id+" IS "+strconv.Itoa(score))
    }

    //errc := make(chan error, 1)
    //go cp(a, b, errc)
    //go cp(b, a, errc)
    //if err := <-errc; err != nil {
    //    log.Println(err)
    //}
    a.Close()
    b.Close()
}

func cp(w io.Writer, r io.Reader, errc chan<- error) {
    _, err := io.Copy(w, r)
    errc <- err
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    base := filepath.Base(path)
    isfile, _ := filepath.Match("*.*", base)
    if !isfile {
        base = ""
    }

    fmt.Println("[http] serving "+path)

    http.ServeFile(w, r, "./"+base)
}
