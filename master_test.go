package etcdlock

import "testing"
import "log"
import "time"
import "math/rand"

var _ = log.Println

func TestLeader(t *testing.T) {
    // Create an etcd client and a lock.
    go startNode("node1")
    go startNode("node2")
    startNode("node3")
}

func startNode(name string) error {
    for {
        client := NewEtcdRegistry([]string{"http://42.121.255.5:4001"})
        lock, err := NewMaster(client, "pushtick", name, 5)
        if err != nil {
            log.Println(err)
            continue
        }
        lock.Start()
        // Start a go routine to process events.
        go processEvents(name, lock.EventsChan())
        // Start the attempt to acquire the lock.
        randtime := time.Duration(rand.Intn(10))
        //log.Println("sleep randtime->", randtime)
        time.Sleep(randtime * time.Second)
        lock.Stop()
    }
}

func processEvents(name string, eventsCh <-chan MasterEvent) {
     for {
        isstop := false
        select {
        case e,ok := <-eventsCh:
            if !ok {
                log.Println(name, "->processEvent stop")
                isstop = true
                break
            }
            if e.Type == MasterAdded {
                log.Println(name, "->Acquired the lock.")
            } else if e.Type == MasterDeleted {
                log.Println(name, "->Lost the lock.")
            } else {
                log.Println(name, "->Lock ownership changed.")
            }
        }
        if isstop {
            break
        }
     }
}

func TestClient(t *testing.T) {
    client := NewEtcdRegistry([]string{"http://42.121.255.5:4001"})
    resp, err := client.Set("key1", "value1", 0)
    if err != nil {
        t.Error(err)
        return
    }
    log.Println(resp.Node)

    resp, err = client.Set("key2", "value2", 0)
    if err != nil {
        t.Error(err)
        return
    }
    log.Println(resp.Node)

    resp, err = client.Get("key1", true, true)
    if err != nil {
        t.Error(err)
        return
    }
    log.Println(resp.Node)
}
