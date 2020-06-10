Feature: String Extraction

  Scenario: Simple print() call
    Given a binary built from source file main.go:
    """
    package main

    func main() {
      print("banana")
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:4       | main.main         |

  Scenario: Simple print() call with constant
    Given a binary built from source file main.go:
    """
    package main

    const x = "banana"

    func main() {
      print(x)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:6       | main.main         |

  Scenario: Simple fmt.Println() call
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    func main() {
      fmt.Println("banana")
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:6       | main.main         |

  Scenario: Simple fmt.Println() call with constant
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    const x = "banana"

    func main() {
      fmt.Println(x)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:8       | main.main         |

  Scenario: Local function call (latter reference to `banana` currently not found)
    Given a binary built from source file main.go:
    """
    package main

    func main() {
      banana := doubleBanana("banana")
      apple := doubleBanana("apple")
      print(banana)
      print(apple)
    }
    
    func doubleBanana(str string) string {
      if str == "banana" {
        return str + "-" + str
      }
      return str
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:4       | main.main         |
      | apple  | main.go:5       | main.main         |

  Scenario: String into struct
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    type foo struct {
      str string
    }

    func main() {
      fmt.Println(&foo{
        str: "banana",
      })
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:11      | main.main         |

  Scenario: String into struct (2)
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    type foo struct {
      str string
    }

    func main() {
      x := &foo{}
      x.str = "banana"
      fmt.Println(x)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:11      | main.main         |

  Scenario: String into struct (3)
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    type foo struct {
      str string
    }

    func main() {
      x := &foo{"banana"}
      fmt.Println(x)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:10      | main.main         |

  Scenario: String into struct (4)
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    type foo struct {
      str1 string
      str2 string
    }

    func main() {
      x := &foo{
        str1: "banana",
        str2: "apple",
      }
      fmt.Println(x)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:12      | main.main         |
      | apple  | main.go:13      | main.main         |

  Scenario: Const into struct
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    const bn = "banana"

    type foo struct {
      str string
    }

    func main() {
      x := &foo{
        str: bn,
      }
      fmt.Println(x)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:13      | main.main         |

  Scenario: Mixed into struct (1)
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    const bn = "banana"

    type foo struct {
      str1 string
      str2 string
    }

    func main() {
      x := &foo{
        str1: bn,
        str2: "apple",
      }
      fmt.Println(x)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:14      | main.main         |
      | apple  | main.go:15      | main.main         |

  Scenario: Mixed into struct (2)
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    const bn = "banana"

    type foo struct {
      str1 string
      str2 string
    }

    func main() {
      x := &foo{
        str1: "apple",
        str2: bn,
      }
      fmt.Println(x)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:15      | main.main         |
      | apple  | main.go:14      | main.main         |

  Scenario: Slice
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    func main() {
      s := []string{
        "banana",
        "apple",
      }
      fmt.Println(s)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:7       | main.main         |
      | apple  | main.go:8       | main.main         |

  Scenario: Slice (2) - PCLN gets this wrong
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    const bn = "banana"

    func main() {
      s := []string{
        bn,
        "apple",
      }
      fmt.Println(s)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:8       | main.main         |
      | apple  | main.go:10      | main.main         |

  Scenario: Slice (3) - PCLN gets this wrong
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    const apl = "apple"

    func main() {
      s := []string{
        "banana",
        apl,
      }
      fmt.Println(s)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:9       | main.main         |
      | apple  | main.go:9       | main.main         |

  Scenario: Struct slice
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    type foo struct {
      str string
    }

    func main() {
      s := []foo{
        {"banana"},
        {"apple"},
      }
      fmt.Println(s)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:11      | main.main         |
      | apple  | main.go:12      | main.main         |

  @wip
  Scenario: Struct repeated value
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    type foo struct {
      str string
    }

    func main() {
      s := []foo{
        {"banana"},
        {"banana"},
      }
      fmt.Println(s)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References    | Symbol References   |
      | banana | main.go:11 main:12 | main.main main.main |

  Scenario: Separate function
    Given a binary built from source file main.go:
    """
    package main

    import "fmt"

    type foo struct {
      str string
    }

    func main() {
      x := createFoo()
      fmt.Println(x)
    }

    func createFoo() *foo {
      return &foo{"banana"}
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:15      | main.createFoo    |

  @wip
  Scenario: String comparison
    Given a binary built from source file main.go:
    """
    package main

    import (
      "os"
      "fmt"
    )

    func main() {
      fmt.Println(isBanana(os.Args[1]))
    }

    func isBanana(str string) bool {
      return str == "banana"
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:10      | main.isBanana     |

  Scenario: String concatenation
    Given a binary built from source file main.go:
    """
    package main

    import (
      "os"
      "fmt"
    )

    func main() {
      fmt.Println(os.Args[1] + "banana")
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:9       | main.main         |

  Scenario: Multiple references
    Given a binary built from source file main.go:
    """
    package main

    import (
      "os"
      "fmt"
    )

    func main() {
      fmt.Println("banana")

      if len(os.Args[1]) > 0 {
        foo()
      }
    }

    func foo() {
      fmt.Println("banana")
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References      | Symbol References  |
      | banana | main.go:9 main.go:17 | main.main main.foo |

  Scenario: Multiple references (const)
    Given a binary built from source file main.go:
    """
    package main

    import (
      "os"
      "fmt"
    )

    const bn = "banana"

    func main() {
      fmt.Println(bn)

      if len(os.Args[1]) > 0 {
        foo()
      }
    }

    func foo() {
      fmt.Println(bn)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References       | Symbol References  |
      | banana | main.go:11 main.go:19 | main.main main.foo |

  @wip
  Scenario: String comparison
    Given a binary built from source file main.go:
    """
    package main

    import (
      "os"
      "fmt"
    )

    func main() {
      fmt.Println(isBanana(os.Args[1]))
    }

    func isBanana(str string) bool {
      if str == "banana" {
        return true
      }
      return false
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:13      | main.isFruit      |

  Scenario: Suffix check
    Given a binary built from source file main.go:
    """
    package main

    import (
      "os"
      "fmt"
      "strings"
    )

    func main() {
      fmt.Println(isBanana(os.Args[1]))
    }

    func isBanana(str string) bool {
      return strings.HasSuffix(str, "banana")
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:14      | main.isBanana     |

  @wip
  Scenario: Suffix check (2)
    Given a binary built from source file main.go:
    """
    package main

    import (
      "os"
      "fmt"
      "strings"
    )

    func main() {
      fmt.Println(isBanana(os.Args[1]))
    }

    func isBanana(str string) (bool, error) {
      if strings.HasSuffix(str, "banana") {
        return true, nil
      }
      return false, fmt.Errorf("no banana: %s", str)
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String        | File References | Symbol References |
      | banana        | main.go:14      | main.isBanana     |
      | no banana: %s | main.go:17      | main.isBanana     |

  Scenario: Mixed
    Given a binary built from source file main.go:
    """
    package main

    import (
      "errors"
      "flag"
      "fmt"
    )

    type Foo struct {
      bananaPrefix string
      bananaScheme string
      ignoreApple  bool
    }

    func main() {
      if err := run(); err != nil {
        fmt.Println(err)
      }
    }

    func run() error {
      var prefix = flag.String("banana-prefix", "", "prefix of the banana")
      var scheme = flag.String("banana-scheme", "", "banana scheme to use")
      var ignoreApple = flag.Bool("ignore-apple", false, "ignore all apples")

      flag.Parse()

      return validate(&Foo{
        bananaPrefix: *prefix,
        bananaScheme: *scheme,
        ignoreApple:  *ignoreApple,
      })
    }

    func validate(f *Foo) error {
      if len(f.bananaPrefix) == 0 || len(f.bananaScheme) == 0 || !f.ignoreApple {
        return errors.New("invalid Foo")
      }
      return nil
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String               | File References | Symbol References |
      | banana-prefix        | main.go:22      | main.run          |
      | prefix of the banana | main.go:22      | main.run          |
      | banana-scheme        | main.go:23      | main.run          |
      | banana scheme to use | main.go:23      | main.run          |
      | ignore-apple         | main.go:24      | main.run          |
      | ignore all apples    | main.go:24      | main.run          |
      | invalid Foo          | main.go:37      | main.validate     |

  Scenario: Panic
    Given a binary built from source file main.go:
    """
    package main

    func main() {
      panic("banana")
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:4       | main.main         |

  Scenario: Panic append
    Given a binary built from source file main.go:
    """
    package main

    import "os"

    func main() {
      panic("banana" + os.Args[1])
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:6       | main.main         |

  Scenario: Panic append (2)
    Given a binary built from source file main.go:
    """
    package main

    import "os"

    func main() {
      panic("banana" + os.Args[1] + "apple")
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:6       | main.main         |
      | apple  | main.go:6       | main.main         |

  Scenario: Panic append (3)
    Given a binary built from source file main.go:
    """
    package main

    import "os"

    func main() {
      panic(os.Args[1] + "banana")
    }
    """
    When that binary is analysed
    Then the following results are returned:
      | String | File References | Symbol References |
      | banana | main.go:6       | main.main         |
