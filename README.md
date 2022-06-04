# Set

Set is a package for comparable generic and hashable interface sets with a few operations.

#### Note: Currently there is a bug in Go 1.18 that causes the Add, Remove, and Contains Hashable set operations to panic. The bug is resolved in the Go dev branch but is currently unreleased.

## Installation
```bash
go get github.com/kastolars/set
```

## Usage

```go
/// Comparable types

// Instantiate
s := New(1, 2, 3, 4)

// Add to set
s.Add(5)

// Remove from set
s.Remove(1)

// Will return false
s.Contains(1)
```

```go
/// Hashable structs

// Give your struct a generic type
type HashableExample[T comparable] struct {
    a, b int
}

// Implement the Hashable interface function "Hash"
func (he HashableExample[T]) Hash() {
    return he.a + he.b
}

he := HashableExample[int]{1, 2}
s := NewHashable[int](th)

// All set operations function the same as Comparable sets
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](https://choosealicense.com/licenses/mit/)