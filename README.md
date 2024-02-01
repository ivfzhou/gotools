##### 通用golang函数库

```shell
go get gitee.com/ivfzhou/gotools/v4@latest
```

```golang
// NewRunner 该函数提供同时最多运行max个协程fn，一旦fn发生error便终止fn运行。
//
// max小于等于0表示不限制协程数。
//
// 朝返回的run函数中添加fn，若block为true表示正在运行的任务数已达到max则会阻塞。
//
// run函数返回error为任务fn返回的第一个error，与wait函数返回的error为同一个。
//
// 注意请在add完所有任务后调用wait。
func NewRunner[T any](ctx context.Context, max int, fn func(context.Context, T) error) (
run func(t T, block bool) error, wait func(fastExit bool) error)

// RunPipeline 将每个jobs依次递给steps函数处理。一旦某个step发生error或者panic，立即返回该error，并及时结束其他协程。
// 除非stopWhenErr为false，则只是终止该job往下一个step投递。
//
// 一个job最多在一个step中运行一次，且一个job一定是依次序递给steps，前一个step处理完毕才会给下一个step处理。
//
// 每个step并发运行jobs。
//
// 等待所有jobs处理结束时会close successCh、errCh，或者ctx被cancel时也将及时结束开启的goroutine后返回。
//
// 从successCh和errCh中获取成功跑完所有step的job和是否发生error。
//
// 若steps中含有nil将会panic。
func RunPipeline[T any](ctx context.Context, jobs []T, stopWhenErr bool, steps ...func(context.Context, T) error) (
successCh <-chan T, errCh <-chan error)

// NewWriteAtReader 获取一个WriterAt和Reader对象，其中WriterAt用于并发写入数据，而与此同时Reader对象同时读取出已经写入好的数据。
//
// WriterAt写入完毕后调用Close，则Reader会全部读取完后结束读取。
//
// WriterAt发生的error会传递给Reader返回。
//
// 该接口是特定为一个目的实现————服务器分片下载数据中转给客户端下载，提高中转数据效率。
func NewWriteAtReader() (WriteAtCloser, io.ReadCloser)

// NewMultiReadCloserToReader 依次从rc中读出数据直到io.EOF则close rc。从r获取rc中读出的数据。
//
// add添加rc，返回error表明读取rc发生错误，可以安全的添加nil。调用endAdd表明不会再有rc添加，当所有数据读完了时，r将返回EOF。
//
// 如果ctx被cancel，将停止读取并返回error。
//
// 所有添加进去的io.ReadCloser都会被close。
func NewMultiReadCloserToReader(ctx context.Context, rc ...io.ReadCloser) (
r io.Reader, add func(rc io.ReadCloser) error, endAdd func())
```

联系电邮：ivfzhou@126.com
