##### 通用golang函数库

```golang
// RunParallel 该函数提供同时运行max个协程fn，一旦fn有err返回则停止接下来的fn运行。
// 朝返回的 add 函数中添加任务，若正在运行的任务数已达到max则会阻塞当前程序。
// add 函数返回err为任务 fn 返回的第一个err。与 wait 函数返回的err为同一个。
// 注意请在 add 完所有任务后调用 wait。
func RunParallel[T any](max int, fn func(T) error) (add func(T) error, wait func() error)

// RunParallelNoBlock 该函数提供同RunParallel一样，但是add函数不会阻塞。注意请在add完所有任务后调用wait。
func RunParallelNoBlock[T any](max int, fn func(T) error) (add func(T) error, wait func() error)

// Run 并发将jobs传递给proc函数运行，一旦发生error便立即返回该error，并结束其它协程。
// 当ctx被cancel时也将立即返回，此时返回cancel时的error。
// 当proc运行发生panic将立即返回该panic字符串化的error。
// proc为nil时函数将panic。
func Run[T any](ctx context.Context, proc func(context.Context, T) error, jobs ...T) error

// StartProcess 将每个jobs依次递给steps函数处理。一旦某个step发生error或者panic，StartProcess立即返回该error，
// 并及时结束其他StartProcess开启的goroutine，也不开启新的goroutine运行step。
// 一个job最多在一个step中运行一次，且一个job一定是依次序递给steps，前一个step处理完毕才会给下一个step处理。
// 每个step并发运行jobs。
// StartProcess等待所有goroutine运行结束才返回，或者ctx被cancel时也将及时结束开启的goroutine后返回。
// StartProcess因被ctx cancel而结束时函数返回nil。若steps中含有nil StartProcess将会panic。
func StartProcess[T any](ctx context.Context, jobs []T, steps ...func(context.Context, T) error) error

// Listen 监听chans，一旦有一个chan激活便立即将T发送给ch，并close ch。
// 若所有chan都未曾激活（chan是nil也认为未激活）且都close了，或者ctx被cancel了，则ch被close。
// 若同时chan被激活和ctx被cancel，则随机返回一个激活发送给chan的值。
func Listen[T any](ctx context.Context, chans ...<-chan T) (ch <-chan T)

// WriteAtReader 获取一个WriterAt和Reader对象，其中WriterAt用于并发写入数据，而与此同时Reader对象同时读取出已经写入好的数据。
// WriterAt写入完毕后调用Close，则Reader会全部读取完后结束读取。
// WriterAt发生的error会传递给Reader返回。
// 该接口是特定为一个目的实现————服务器分片下载数据中转给客户端下载，提高中转数据效率。
func WriteAtReader() (WriteAtCloser, io.ReadCloser)

// NewMultiReader 依次从reader读出数据并写入writer中，并close reader。
// 返回send用于添加reader，readSize表示需要从reader读出的字节数，order用于表示记录读取序数并传递给writer，若读取字节数对不上则返回error。
// 返回wait用于等待所有reader读完，若读取发生error，wait返回该error，并结束读取。
// 务必等所有reader都已添加给send后再调用wait。
// 该函数可用于需要非同一时间多个读取流和一个写入流的工作模型。
func NewMultiReader(ctx context.Context, writer func(order int, p []byte)) (
send func(readSize, order int, reader io.ReadCloser), wait func() error)

// MultiReadCloser 依次从rc中读出数据直到io.EOF则close rc。从r获取rc中读出的数据。
// add添加rc，返回error表明读取rc发生错误，可以安全的添加nil。调用endAdd表明不会再有rc添加，当所有数据读完了时，r将返回EOF。
// 如果ctx被cancel，将停止读取并返回error。
// 所有添加进去的io.ReadCloser都会被close。
func MultiReadCloser(ctx context.Context, rc ...io.ReadCloser) (r io.Reader, add func(rc io.ReadCloser) error, endAdd func())
```

联系电邮：ivfzhou@126.com
