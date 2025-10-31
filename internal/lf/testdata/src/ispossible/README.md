# IsPossibleConverter Test Cases

This directory contains comprehensive test cases for the `IsPossibleConverter` function, which determines whether a function is likely a converter function based on its signature.

## Test Structure

- **models.go** - Type definitions used in test cases
- **positive_cases.go** - Functions that SHOULD be detected as converters
- **negative_cases.go** - Functions that should NOT be detected as converters

## Positive Test Cases (12 cases)

These functions should be detected as converters:

1. **ConvertUserToDTO** - Basic struct-to-struct conversion
2. **ConvertUserPtrToDTO** - Pointer input conversion
3. **ConvertUserToDTOPtr** - Pointer output conversion
4. **ConvertUserPtrToDTOPtr** - Pointer to pointer conversion
5. **ConvertUsersToDTO** - Slice conversion
6. **ConvertUserSlicePtrToDTO** - Slice of pointers conversion
7. **ConvertProductToResponse** - Different suffix pattern (Product→ProductResponse)
8. **ConvertProductMap** - Map conversion
9. **TransformUserToDTO** - Transform naming convention
10. **UserToDTO** - Short naming convention
11. **ToUserDTO** - Even shorter naming convention
12. **BuildUserDTOFromUser** - Builder-style naming

## Negative Test Cases (13 cases)

These functions should NOT be detected as converters:

1. **NoParams** - Function with no parameters
2. **NoResults** - Function with no return values
3. **NoStructParams** - Only primitive type parameters
4. **UnrelatedTypes** - Struct params but no naming similarity
5. **SameTypeInOut** - Same type as input and output (TODO: known bug)
6. **MultipleUnrelatedStructs** - Multiple struct params, no clear conversion pair
7. **OnlyPrimitiveReturn** - Returns only primitives despite struct input
8. **OnlyErrorReturn** - Returns only error
9. **SliceToNonSlice** - Incompatible container types
10. **NonSliceToSlice** - Incompatible container types
11. **MapToSlice** - Incompatible container types
12. **SliceToMap** - Incompatible container types
13. **HelperFunction** - Utility function with similar types but different purpose

## Special Cases

### WithContextAndError
This function returns `(UserDTO, error)` and IS detected as a converter. This is correct behavior since many converter functions return errors for validation purposes.

### SameTypeInOut (Known Issue)
Currently commented out in tests. This is a known bug mentioned in `analyzer.go:168` - the function incorrectly identifies same-type input/output as converters.

## Running Tests

```bash
go test -v ./internal/lf -run TestIsPossibleConverter
```

## Coverage

The test suite covers:
- ✅ Basic struct-to-struct conversions
- ✅ Pointer conversions (all combinations)
- ✅ Container types (slices, maps)
- ✅ Various naming conventions
- ✅ Container type compatibility rules
- ✅ Edge cases (no params, no results, primitives only)
- ✅ False positives (helper functions, unrelated types)
- ⚠️ Same-type conversions (known bug, not yet fixed)
