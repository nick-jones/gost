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
