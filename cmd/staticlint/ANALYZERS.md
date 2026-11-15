# staticlint analyzer catalog

The registry is built from Staticcheck v0.8.2. These descriptions are the analyzer documentation shipped with that pinned release. The executable registers the complete SA family and every ST style check. The [package documentation](doc.go) describes the standard Go analyzers, the project-specific rule, and the two additional public analyzers.

## Staticcheck SA analyzers

These checks detect correctness bugs, suspicious API use, concurrency mistakes, and performance problems.

| ID | Purpose |
| --- | --- |
| [SA1000](https://staticcheck.dev/docs/checks/#sa1000) | Invalid regular expression |
| [SA1001](https://staticcheck.dev/docs/checks/#sa1001) | Invalid template |
| [SA1002](https://staticcheck.dev/docs/checks/#sa1002) | Invalid format in time.Parse |
| [SA1003](https://staticcheck.dev/docs/checks/#sa1003) | Unsupported argument to functions in encoding/binary |
| [SA1004](https://staticcheck.dev/docs/checks/#sa1004) | Suspiciously small untyped constant in time.Sleep |
| [SA1005](https://staticcheck.dev/docs/checks/#sa1005) | Invalid first argument to exec.Command |
| [SA1006](https://staticcheck.dev/docs/checks/#sa1006) | Printf with dynamic first argument and no further arguments |
| [SA1007](https://staticcheck.dev/docs/checks/#sa1007) | Invalid URL in net/url.Parse |
| [SA1008](https://staticcheck.dev/docs/checks/#sa1008) | Non-canonical key in http.Header map |
| [SA1010](https://staticcheck.dev/docs/checks/#sa1010) | (*regexp.Regexp).FindAll called with n == 0, which will always return zero results |
| [SA1011](https://staticcheck.dev/docs/checks/#sa1011) | Various methods in the 'strings' package expect valid UTF-8, but invalid input is provided |
| [SA1012](https://staticcheck.dev/docs/checks/#sa1012) | A nil context.Context is being passed to a function, consider using context.TODO instead |
| [SA1013](https://staticcheck.dev/docs/checks/#sa1013) | io.Seeker.Seek is being called with the whence constant as the first argument, but it should be the second |
| [SA1014](https://staticcheck.dev/docs/checks/#sa1014) | Non-pointer value passed to Unmarshal or Decode |
| [SA1015](https://staticcheck.dev/docs/checks/#sa1015) | Using time.Tick in a way that will leak. Consider using time.NewTicker, and only use time.Tick in tests, commands and endless functions |
| [SA1016](https://staticcheck.dev/docs/checks/#sa1016) | Trapping a signal that cannot be trapped |
| [SA1017](https://staticcheck.dev/docs/checks/#sa1017) | Channels used with os/signal.Notify should be buffered |
| [SA1018](https://staticcheck.dev/docs/checks/#sa1018) | strings.Replace called with n == 0, which does nothing |
| [SA1019](https://staticcheck.dev/docs/checks/#sa1019) | Using a deprecated function, variable, constant or field |
| [SA1020](https://staticcheck.dev/docs/checks/#sa1020) | Using an invalid host:port pair with a net.Listen-related function |
| [SA1021](https://staticcheck.dev/docs/checks/#sa1021) | Using bytes.Equal to compare two net.IP |
| [SA1023](https://staticcheck.dev/docs/checks/#sa1023) | Modifying the buffer in an io.Writer implementation |
| [SA1024](https://staticcheck.dev/docs/checks/#sa1024) | A string cutset contains duplicate characters |
| [SA1025](https://staticcheck.dev/docs/checks/#sa1025) | It is not possible to use (*time.Timer).Reset's return value correctly |
| [SA1026](https://staticcheck.dev/docs/checks/#sa1026) | Cannot marshal channels or functions |
| [SA1027](https://staticcheck.dev/docs/checks/#sa1027) | Atomic access to 64-bit variable must be 64-bit aligned |
| [SA1028](https://staticcheck.dev/docs/checks/#sa1028) | sort.Slice can only be used on slices |
| [SA1029](https://staticcheck.dev/docs/checks/#sa1029) | Inappropriate key in call to context.WithValue |
| [SA1030](https://staticcheck.dev/docs/checks/#sa1030) | Invalid argument in call to a strconv function |
| [SA1031](https://staticcheck.dev/docs/checks/#sa1031) | Overlapping byte slices passed to an encoder |
| [SA1032](https://staticcheck.dev/docs/checks/#sa1032) | Wrong order of arguments to errors.Is |
| [SA2000](https://staticcheck.dev/docs/checks/#sa2000) | (*sync.WaitGroup).Add called inside the goroutine, leading to a race condition |
| [SA2001](https://staticcheck.dev/docs/checks/#sa2001) | Empty critical section, did you mean to defer the unlock? |
| [SA2002](https://staticcheck.dev/docs/checks/#sa2002) | Called testing.T.FailNow or SkipNow in a goroutine, which isn't allowed |
| [SA2003](https://staticcheck.dev/docs/checks/#sa2003) | Deferred Lock right after locking, likely meant to defer Unlock instead |
| [SA3000](https://staticcheck.dev/docs/checks/#sa3000) | TestMain doesn't call os.Exit, hiding test failures |
| [SA3001](https://staticcheck.dev/docs/checks/#sa3001) | Assigning to b.N in benchmarks distorts the results |
| [SA4000](https://staticcheck.dev/docs/checks/#sa4000) | Binary operator has identical expressions on both sides |
| [SA4001](https://staticcheck.dev/docs/checks/#sa4001) | &*x gets simplified to x, it does not copy x |
| [SA4003](https://staticcheck.dev/docs/checks/#sa4003) | Comparing unsigned values against negative values is pointless |
| [SA4004](https://staticcheck.dev/docs/checks/#sa4004) | The loop exits unconditionally after one iteration |
| [SA4005](https://staticcheck.dev/docs/checks/#sa4005) | Field assignment that will never be observed. Did you mean to use a pointer receiver? |
| [SA4006](https://staticcheck.dev/docs/checks/#sa4006) | A value assigned to a variable is never read before being overwritten. Forgotten error check or dead code? |
| [SA4008](https://staticcheck.dev/docs/checks/#sa4008) | The variable in the loop condition never changes, are you incrementing the wrong variable? |
| [SA4009](https://staticcheck.dev/docs/checks/#sa4009) | A function argument is overwritten before its first use |
| [SA4010](https://staticcheck.dev/docs/checks/#sa4010) | The result of append will never be observed anywhere |
| [SA4011](https://staticcheck.dev/docs/checks/#sa4011) | Break statement with no effect. Did you mean to break out of an outer loop? |
| [SA4012](https://staticcheck.dev/docs/checks/#sa4012) | Comparing a value against NaN even though no value is equal to NaN |
| [SA4013](https://staticcheck.dev/docs/checks/#sa4013) | Negating a boolean twice (!!b) is the same as writing b. This is either redundant, or a typo. |
| [SA4014](https://staticcheck.dev/docs/checks/#sa4014) | An if/else if chain has repeated conditions and no side-effects; if the condition didn't match the first time, it won't match the second time, either |
| [SA4015](https://staticcheck.dev/docs/checks/#sa4015) | Calling functions like math.Ceil on floats converted from integers doesn't do anything useful |
| [SA4016](https://staticcheck.dev/docs/checks/#sa4016) | Certain bitwise operations, such as x ^ 0, do not do anything useful |
| [SA4017](https://staticcheck.dev/docs/checks/#sa4017) | Discarding the return values of a function without side effects, making the call pointless |
| [SA4018](https://staticcheck.dev/docs/checks/#sa4018) | Self-assignment of variables |
| [SA4019](https://staticcheck.dev/docs/checks/#sa4019) | Multiple, identical build constraints in the same file |
| [SA4020](https://staticcheck.dev/docs/checks/#sa4020) | Unreachable case clause in a type switch |
| [SA4021](https://staticcheck.dev/docs/checks/#sa4021) | 'x = append(y)' is equivalent to 'x = y' |
| [SA4022](https://staticcheck.dev/docs/checks/#sa4022) | Comparing the address of a variable against nil |
| [SA4023](https://staticcheck.dev/docs/checks/#sa4023) | Impossible comparison of interface value with untyped nil |
| [SA4024](https://staticcheck.dev/docs/checks/#sa4024) | Checking for impossible return value from a builtin function |
| [SA4025](https://staticcheck.dev/docs/checks/#sa4025) | Integer division of literals that results in zero |
| [SA4026](https://staticcheck.dev/docs/checks/#sa4026) | Go constants cannot express negative zero |
| [SA4027](https://staticcheck.dev/docs/checks/#sa4027) | (*net/url.URL).Query returns a copy, modifying it doesn't change the URL |
| [SA4028](https://staticcheck.dev/docs/checks/#sa4028) | x % 1 is always zero |
| [SA4029](https://staticcheck.dev/docs/checks/#sa4029) | Ineffective attempt at sorting slice |
| [SA4030](https://staticcheck.dev/docs/checks/#sa4030) | Ineffective attempt at generating random number |
| [SA4031](https://staticcheck.dev/docs/checks/#sa4031) | Checking never-nil value against nil |
| [SA4032](https://staticcheck.dev/docs/checks/#sa4032) | Comparing runtime.GOOS or runtime.GOARCH against impossible value |
| [SA5000](https://staticcheck.dev/docs/checks/#sa5000) | Assignment to nil map |
| [SA5001](https://staticcheck.dev/docs/checks/#sa5001) | Deferring Close before checking for a possible error |
| [SA5002](https://staticcheck.dev/docs/checks/#sa5002) | The empty for loop ('for {}') spins and can block the scheduler |
| [SA5003](https://staticcheck.dev/docs/checks/#sa5003) | Defers in infinite loops will never execute |
| [SA5004](https://staticcheck.dev/docs/checks/#sa5004) | 'for { select { ...' with an empty default branch spins |
| [SA5005](https://staticcheck.dev/docs/checks/#sa5005) | The finalizer references the finalized object, preventing garbage collection |
| [SA5007](https://staticcheck.dev/docs/checks/#sa5007) | Infinite recursive call |
| [SA5008](https://staticcheck.dev/docs/checks/#sa5008) | Invalid struct tag |
| [SA5009](https://staticcheck.dev/docs/checks/#sa5009) | Invalid Printf call |
| [SA5010](https://staticcheck.dev/docs/checks/#sa5010) | Impossible type assertion |
| [SA5012](https://staticcheck.dev/docs/checks/#sa5012) | Passing odd-sized slice to function expecting even size |
| [SA6000](https://staticcheck.dev/docs/checks/#sa6000) | Using regexp.Match or related in a loop, should use regexp.Compile |
| [SA6001](https://staticcheck.dev/docs/checks/#sa6001) | Missing an optimization opportunity when indexing maps by byte slices |
| [SA6002](https://staticcheck.dev/docs/checks/#sa6002) | Storing non-pointer values in sync.Pool allocates memory |
| [SA6003](https://staticcheck.dev/docs/checks/#sa6003) | Converting a string to a slice of runes before ranging over it |
| [SA6005](https://staticcheck.dev/docs/checks/#sa6005) | Inefficient string comparison with strings.ToLower or strings.ToUpper |
| [SA6006](https://staticcheck.dev/docs/checks/#sa6006) | Using io.WriteString to write []byte |
| [SA9001](https://staticcheck.dev/docs/checks/#sa9001) | Defers in range loops may not run when you expect them to |
| [SA9002](https://staticcheck.dev/docs/checks/#sa9002) | Using a non-octal os.FileMode that looks like it was meant to be in octal. |
| [SA9003](https://staticcheck.dev/docs/checks/#sa9003) | Empty body in an if or else branch |
| [SA9004](https://staticcheck.dev/docs/checks/#sa9004) | Only the first constant has an explicit type |
| [SA9005](https://staticcheck.dev/docs/checks/#sa9005) | Trying to marshal a struct with no public fields nor custom marshaling |
| [SA9006](https://staticcheck.dev/docs/checks/#sa9006) | Dubious bit shifting of a fixed size integer value |
| [SA9007](https://staticcheck.dev/docs/checks/#sa9007) | Deleting a directory that shouldn't be deleted |
| [SA9008](https://staticcheck.dev/docs/checks/#sa9008) | else branch of a type assertion is probably not reading the right value |
| [SA9009](https://staticcheck.dev/docs/checks/#sa9009) | Ineffectual Go compiler directive |
| [SA9010](https://staticcheck.dev/docs/checks/#sa9010) | Returned function should be called in defer |

## Staticcheck ST analyzers

These checks enforce Go documentation, naming, error formatting, and style conventions.

| ID | Purpose |
| --- | --- |
| [ST1000](https://staticcheck.dev/docs/checks/#st1000) | Incorrect or missing package comment |
| [ST1001](https://staticcheck.dev/docs/checks/#st1001) | Dot imports are discouraged |
| [ST1003](https://staticcheck.dev/docs/checks/#st1003) | Poorly chosen identifier |
| [ST1005](https://staticcheck.dev/docs/checks/#st1005) | Incorrectly formatted error string |
| [ST1006](https://staticcheck.dev/docs/checks/#st1006) | Poorly chosen receiver name |
| [ST1008](https://staticcheck.dev/docs/checks/#st1008) | A function's error value should be its last return value |
| [ST1011](https://staticcheck.dev/docs/checks/#st1011) | Poorly chosen name for variable of type time.Duration |
| [ST1012](https://staticcheck.dev/docs/checks/#st1012) | Poorly chosen name for error variable |
| [ST1013](https://staticcheck.dev/docs/checks/#st1013) | Should use constants for HTTP error codes, not magic numbers |
| [ST1015](https://staticcheck.dev/docs/checks/#st1015) | A switch's default case should be the first or last case |
| [ST1016](https://staticcheck.dev/docs/checks/#st1016) | Use consistent method receiver names |
| [ST1017](https://staticcheck.dev/docs/checks/#st1017) | Don't use Yoda conditions |
| [ST1018](https://staticcheck.dev/docs/checks/#st1018) | Avoid zero-width and control characters in string literals |
| [ST1019](https://staticcheck.dev/docs/checks/#st1019) | Importing the same package multiple times |
| [ST1020](https://staticcheck.dev/docs/checks/#st1020) | The documentation of an exported function should start with the function's name |
| [ST1021](https://staticcheck.dev/docs/checks/#st1021) | The documentation of an exported type should start with type's name |
| [ST1022](https://staticcheck.dev/docs/checks/#st1022) | The documentation of an exported variable or constant should start with variable's name |
| [ST1023](https://staticcheck.dev/docs/checks/#st1023) | Redundant type in variable declaration |

