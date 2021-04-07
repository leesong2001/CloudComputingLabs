package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

const tot = 100000
const threadNUM = 15
const maxn = 81*4*81 + 10
const maxr = 9*9*9 + 10
const maxc = 81*4 + 10

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
func (x *DLX) Solve(inp string, c2 chan info, p int) {
	x.init(9, inp)
	if !x.dfs(0) {
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
	c2 <- info{string(res), p}
}

type info struct {
	str string
	p   int
}

func thwk(c1, c2 chan info) {
	var llx DLX
	for {
		x := <-c1
		llx.Solve(x.str, c2, x.p)
	}
}
func mainwork(fl, fot *os.File) {
	now := time.Now()
	re := bufio.NewReader(fl)
	ot := bufio.NewWriter(fot)
	c1 := make(chan info, tot)
	c2 := make(chan info, tot)
	var ans [tot]string
	for i := 0; i < threadNUM; i++ {
		go thwk(c1, c2)
	}
	var x string
	cnt := 0
	for {
		_, err := fmt.Fscan(re, &x)
		if err != nil && err.Error() == io.EOF.Error() {
			break
		}
		c1 <- info{x, cnt}
		cnt++
	}
	for i := 0; i < cnt; i++ {
		x := <-c2
		ans[x.p] = x.str
	}
	//fmt.Println(time.Since(now))
	for j := 0; j < cnt; j++ {
		fmt.Fprintln(ot, ans[j])
	}
	fmt.Fprintln(ot, time.Since(now))
	ot.Flush()
	fl.Close()
	fot.Close()
}
func main() {
	for {
		fmt.Printf("input a filename or exit: ")
		var flname string
		fmt.Scan(&flname)
		if flname == "exit" {
			break
		}
		f, err := os.Open(flname)
		if err != nil {
			fmt.Println("file not exist")
		} else {
			fout, _ := os.Create(flname + ".out")
			go mainwork(f, fout)
		}
	}
	fmt.Println("program has terminated")
}
