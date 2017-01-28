//Adding the RTMP server.  This will put up a RTMP endpoint when starting up Swarm.
//It's a simple RTMP server that will take a video stream and play it right back out.
//After bringing up the Swarm node with RTMP enabled, try it out using:
//
//ffmpeg -re -i bunny.mp4 -c copy -f flv rtmp://localhost/movie
//ffplay rtmp://localhost/movie

package rtmp

import (
    "github.com/nareix/joy4/av/avutil"
    "github.com/nareix/joy4/av/pubsub"
    "github.com/nareix/joy4/format"
    "github.com/nareix/joy4/format/rtmp"
    "sync"
)

func init() {
    format.RegisterAll()
}

func StartRtmpServer() {
    server := &rtmp.Server{}

    l := &sync.RWMutex{}
    type Wrapper struct {
        que *pubsub.Queue
    }

    channels := map[string]*Wrapper{}

    server.HandlePlay = func(conn *rtmp.Conn) {
        l.RLock()
        ch := channels[conn.URL.Path]
        l.RUnlock()

        if ch != nil {
            cursor := ch.que.Latest()
            avutil.CopyFile(conn, cursor)
        }
    }

    server.HandlePublish = func(conn *rtmp.Conn) {
        l.Lock()
        ch := channels[conn.URL.Path]

        if ch == nil {
            ch = &Wrapper{}
            ch.que = pubsub.NewQueue()
            channels[conn.URL.Path] = ch
        } else {
            ch = nil
        }
        l.Unlock()
        if (ch == nil) {
            return
        }

        avutil.CopyFile(ch.que, conn)

        l.Lock()
        delete(channels, conn.URL.Path)
        l.Unlock()
        ch.que.Close()
    }

    server.ListenAndServe()
}

