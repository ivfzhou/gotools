##### 常用函数库

```golang
// StartProcess 分阶段依次运行steps函数处理jobs数据。每个阶段并发运行，每个job将会依次给steps处理，
// 即一个job前面step运行完毕，后面step才会开始运行该job。
// 一当step发生error或者panic StartProcess便立即返回，且结束其他所有开启的协程。
// 当ctx被cancel时也将立即返回，此时返回context.Cause()的error。
func StartProcess[T any](ctx context.Context, jobs []T, steps ...func(context.Context, T) error)

// Run 并发将jobs传递给proc函数运行，一旦发生error便立即返回该error，并结束其它协程。
// 当ctx被cancel时也将立即返回，此时返回cancel时的error。
// 当proc运行发生panic将立即返回该panic字符串化的error。
func Run[T any](ctx context.Context, proc func(context.Context, T) error, jobs ...T) error

// RunParallel 该函数提供同时运行 max 个协程 fn，一旦 fn 有err返回则停止接下来的fn运行。
// 朝返回的 add 函数中添加任务，若正在运行的任务数已达到max则会阻塞当前程序。
// add 函数返回err为任务 fn 返回的第一个err。与 wait 函数返回的err为同一个。
// 注意请在 add 完所有任务后调用 wait。
func RunParallel[T any](max int, fn func(T) error) (add func(T) error, wait func() error)

// RunParallelNoBlock 该函数提供同 RunParallel 一样，但是 add 函数不会阻塞。注意请在 add 完所有任务后调用 wait。
func RunParallelNoBlock[T any](max int, fn func(T) error) (add func(T) error, wait func() error)

// Listen 监听所有chan，一旦有一个chan激活便立即将T发送给函数返回的ch，之后ch被close。
// 若所有chan都未曾激活且都close了，或者ctx被cancel了，则ch被close。
func Listen[T any](ctx context.Context, chans ...<-chan T) <-chan T

```

`RunParallel`
```golang
type Data struct {}
maxRunnings := 12 // 最大同时运行协程数
work := func (data Data) error { // 每个协程处理内容
    println("doing")
}
add, wait := gotools.RunParallel(maxRunning, work)
var datas []*Data // 需要处理的数据
for _, data := range datas {
	err := add(data) // 添加处理，如果达到maxRunnings则阻塞，不阻塞见RunParallelNoBlock
	if err != nil {
		// 说明有协程工作返回err，其它还未运行的协程将取消
    }
}
err := wait() // 等待处理完毕
if err != nil {
	// 返回第一个处理失败的协程返回的err
}
```

`RunConcurrently`
```golang
work1 := func() error{}
work2 := func() error{}
work3 := func() error{}
wait := gotools.RunConcurrently(work1, work2, work3) // 同时运行所有函数
err := wait() // 等待运行完毕
if err != nil {
	// 返回第一个处理失败的err，其它还未运行work将取消运行
}
```

联系电邮：ivfzhou@126.com
