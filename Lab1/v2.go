package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

const tot = 1000
const threadNUM = 20

const maxn = 81*4*81 + 10
const maxr = 9*9*9 + 10
const maxc = 81*4 + 10

var ans [tot]string

//DLX 舞蹈链
type DLX struct {
	n, sz, ansd          int
	s                    [maxc]int
	row, col, l, r, u, d [maxn]int
	vec                  []int
	ans                  [maxr]int
	sudoku               [10][10]int
}

func (x *DLX) addrow(r int) {
	first := x.sz
	for i := 0; i < len(x.vec); i++ {
		c := x.vec[i]
		x.l[x.sz] = x.sz - 1
		x.r[x.sz] = x.sz + 1
		x.d[x.sz] = c
		x.u[x.sz] = x.u[c]
		x.d[x.u[c]] = x.sz
		x.u[c] = x.sz
		x.row[x.sz] = r
		x.col[x.sz] = c
		x.s[c]++
		x.sz++
	}
	x.r[x.sz-1] = first
	x.l[first] = x.sz - 1
}
func (x *DLX) encode(a, b, c int) int {
	return 81*a + b*9 + c + 1
}
func (x *DLX) trans(a, b int) int {
	a /= 3
	b /= 3
	return a*3 + b
}
func (x *DLX) build() {
	x.vec = make([]int, 0)
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			for k := 0; k < 9; k++ {
				if x.sudoku[i][j] == -1 || x.sudoku[i][j] == k {
					x.vec = make([]int, 0)
					x.vec = append(x.vec, x.encode(0, i, j))
					x.vec = append(x.vec, x.encode(1, i, k))
					x.vec = append(x.vec, x.encode(2, j, k))
					x.vec = append(x.vec, x.encode(3, x.trans(i, j), k))
					x.addrow(x.encode(i, j, k))
				}
			}
		}
	}
}

// Init 初始化DLX
func (x *DLX) init(sz int, sudo string) {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			x.sudoku[i][j] = int(sudo[i*9+j]-'0') - 1
		}
	}
	n := sz * sz * 4
	x.n = n
	for i := 0; i <= n; i++ {
		x.u[i], x.d[i] = i, i
		x.l[i], x.r[i] = i-1, i+1
	}
	x.r[n], x.l[0] = 0, n
	x.sz = n + 1
	for i := 0; i < maxc; i++ {
		x.s[i] = 0
	}
	x.build()
}

func (x *DLX) remove(c int) {
	x.l[x.r[c]] = x.l[c]
	x.r[x.l[c]] = x.r[c]
	for i := x.d[c]; i != c; i = x.d[i] {
		for j := x.r[i]; j != i; j = x.r[j] {
			x.u[x.d[j]] = x.u[j]
			x.d[x.u[j]] = x.d[j]
			x.s[x.col[j]]--
		}
	}

}

func (x *DLX) restore(c int) {
	for i := x.u[c]; i != c; i = x.u[i] {
		for j := x.l[i]; j != i; j = x.l[j] {
			x.s[x.col[j]]++
			x.u[x.d[j]] = j
			x.d[x.u[j]] = j
		}
	}
	x.l[x.r[c]] = c
	x.r[x.l[c]] = c
}

func (x *DLX) dfs(d int) bool {
	if x.r[0] == 0 {
		x.ansd = d
		return true
	}
	c := x.r[0]
	for i := x.r[0]; i != 0; i = x.r[i] {
		if x.s[i] < x.s[c] {
			c = i
		}
	}
	x.remove(c)
	for i := x.d[c]; i != c; i = x.d[i] {
		x.ans[d] = x.row[i]
		for j := x.r[i]; j != i; j = x.r[j] {
			x.remove(x.col[j])
		}
		if x.dfs(d + 1) {
			return true
		}
		for j := x.l[i]; j != i; j = x.l[j] {
			x.restore(x.col[j])
		}
	}
	x.restore(c)
	return false
}

func (x *DLX) decode(code int, a, b, c *int) {
	code--
	*c = code % 9
	code /= 9
	*b = code % 9
	code /= 9
	*a = code
}

// Solve 解决数独
func (x *DLX) Solve(inp string, c1 chan string, c2 chan int, p int, cond *sync.Cond) {
	x.init(9, inp)
	if !x.dfs(0) {
		c1 <- ""
		c2 <- p
		return
	}
	var res []byte
	for i := 0; i < x.ansd; i++ {
		var r, c, v int
		x.decode(x.ans[i], &r, &c, &v)
		x.sudoku[r][c] = v
	}
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			res = append(res, byte(x.sudoku[i][j]+'1'))
		}
	}
	cond.L.Lock()
	c1 <- string(res)
	c2 <- p
	cond.L.Unlock()
}

var lx [threadNUM]DLX

func thwk(c1 chan string, c2 chan int, c3 chan string, c4 chan int, cond *sync.Cond, p int) {
	for {
		x := <-c3
		lx[p].Solve(x, c1, c2, <-c4, cond)
	}
}
func mainwork() {
	now := time.Now()
	re := bufio.NewReader(os.Stdin)
	mux := sync.Mutex{}
	cond := sync.NewCond(&mux)

	c1 := make(chan string, tot)
	c2 := make(chan int, tot)
	c3 := make(chan string, tot)
	c4 := make(chan int, tot)
	for i := 0; i < threadNUM; i++ {
		go thwk(c1, c2, c3, c4, cond, i)
	}
	var x string
	for i := 0; i < tot; i++ {
		fmt.Fscan(re, &x)
		c3 <- x
		c4 <- i
	}
	for i := 0; i < tot; i++ {
		ans[<-c2] = <-c1
	}
	for j := 0; j < tot; j++ {
		fmt.Println(ans[j])
	}
	fmt.Println(time.Since(now))
}
func main() {
	mainwork()
}
