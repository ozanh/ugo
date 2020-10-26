# Operators

uGO uses same binary and unary operators with Go. uGO relies on Go's wrap-around
behavior if compiler does not catch any overflow. Floating point numbers may be
rounded due to conversions. It is a good practice to use [builtin conversion
functions](builtins.md) to resolve ambiguities. If optimizer is enabled in
compiler options, using conversion functions with constant values has no
overhead at runtime.

## Binary Operators

### Relational Operators

Relational operators yields bool values.

| Symbol | Operation             |
|:------:|-----------------------|
|   ==   | equal                 |
|   !=   | not equal             |
|    <   | less than             |
|    >   | greater than          |
|   <=   | less than or equal    |
|   >=   | greater than or equal |

All relational operators apply to int, uint, float, char, bool, string, bytes.
uGO's [Object](tutorial.md#interfaces) interface methods `Equal and BinaryOp`
evaluates these operations. A runtime error is thrown if types are not
comparable. Note that, bool values are converted to untyped 1 or 0 for true or
false to compare with numeric values.

Following is the conversion table to apply relational operators to different
numeric types. `p` is the left hand side (LHS) in the first column and `q` is
the right hand side (RHS) in the first row.

| Type      | int        | uint       | float          | char          |
|:----------|------------|------------|----------------|---------------|
| **int**   | -          | uint64(p)  | float64(p)     | rune(p)       |
| **uint**  | uint64(q)  | -          | float64(p)     | rune(p)       |
| **float** | float64(q) | float64(q) | -              | **TypeError** |
| **char**  | rune(q)    | rune(q)    | **TypeError**  | -             |

- For `map` values, `==` and `!=` operators are applicable if LHS and RHS
  operands are of `map` type

- For `array` values, `==` and `!=` operators are applicable if LHS and RHS
  operands are of `array` type

- For `string` and `bytes` values, all relational operators are applicable if
  LHS and RHS are of same type

### Binary Arithmetic Operators

| Symbol | Operation          | Supported Types                              |
|:------:|--------------------|----------------------------------------------|
|    +   | sum                | int, uint, float, char, string, bytes, array |
|    -   | difference         | int, uint, float, char                       |
|    *   | product            | int, uint, float                             |
|    /   | quotient           | int, uint, float                             |
|   %    | remainder          | int, uint                                    |
|   &    | bitwise AND        | int, uint                                    |
|   \|   | bitwise OR         | int, uint                                    |
|   ^    | bitwise XOR        | int, uint                                    |
|   &^   | bit clear (AND NOT)| int, uint                                    |
|   <<   | shift left         | int, uint                                    |
|   >>   | shift right        | int, uint                                    |

**Rules**

LHS: left hand side, RHS: right hand side

- `string` values only support `+` operator for string concatenation as LHS
  operand regardless of other operand's type. Result is always of `string` type
- `bytes` values only support `+` operator for byte concatenation if it is LHS
  operand and RHS is of `bytes` or `string` type
- `array` values only support `+` operator to append object if it is LHS operand
- `bool` values are treated as untyped 1 or 0 before arithmetic operation
- `char` values only support `+`, `-` operators with `char`, `int`, `uint`
  values
- `char` values support `*`, `/`, `%`, `|`, `^`, `&^`, `<<`, `>>` operators
  if both operands are of `char` type
- if LHS or RHS is of `char` type, other operand is converted to `char` value
- if LHS or RHS is of `float` type, other operand is converted to `float` value
- if LHS or RHS is unsigned integer, signed integer is converted to unsigned
  integer
- A runtime error `TypeError` is thrown if operand is not of expected type

## Unary Operators

| Symbol | Operation                  | Supported Types                |
|:------:|----------------------------|--------------------------------|
|    +   | positive `0 + x`           | int, uint, float, char, bool*  |
|    -   | negation `0 - x`           | int, uint, float, char, bool*  |
|    ^   | bitwise complement** `m^x` | int, uint, char, bool*         |
|    !   | logical negation NOT `!x`  | all types                      |
|   ++   | increment `x = x + 1`      | all types except map and error |
|   --   | decrement `x = x - 1`      | all types except map and error |

\* bool values are converted to int 1 or 0 to apply operation

\*\* with m = "all bits set to 1" for unsigned x and  m = -1 for signed x

## Logical Operators

Logical operators apply to all types and yield a result of the same type as the
operands. The right operand is evaluated conditionally.

| Symbol | Operation                                           |
|:------:|-----------------------------------------------------|
|   &&   | Logical AND  `p && q  is  "if p then q else false"` |
|   \|\| | Logical OR  `p \|\| q  is  "if p then true else q"` |

_See [builtins](builtins.md) for conversion and type checking functions_
