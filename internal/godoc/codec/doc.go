// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package codec implements the general-purpose part of an encoder for Go
values. It relies on code generation rather than reflection so it is
significantly faster than reflection-based encoders like gob. It also
preserves sharing among struct pointers (but not other forms of sharing, like
other pointer types or sub-slices). These features are sufficient for
encoding the structures of the go/ast package, which is its sole purpose.

# Encoding Scheme

Every encoded value begins with a single byte that describes what (if anything) follows.
A value's encoding contains enough information to skip over the value, since the
decoder must be able to do that if it encounters a struct field it doesn't know.

Most of the values of that initial byte can be devoted to small unsigned
integers. For example, the number 17 is represented by the single byte 17.
Only a few byte values have special meanings, whose descriptions follow.

The nil code indicates that the value is nil. We don't absolutely need this:
we could always represent the nil value for a type as something that couldn't
be mistaken for an encoded value of that type. For instance, we could use 0
for nil in the case of slices (which always begin with the nValues code), and
for pointers to numbers like *int, we could use something like "nBytes 0".
But it is simpler to have a reserved value for nil.

The nBytes code indicates that an unsigned integer N is encoded next,
followed by N bytes of data. This is used to represent strings and byte
slices, as well numbers bigger than can fit into the initial byte. For
example, the string "hi" is represented as:

	nBytes 2 'h' 'i'

Unsigned integers that can't fit into the initial byte are encoded as byte
sequences of length 4 or 8, holding little-endian uint32 or uint64 values. We
use uint32s where possible to save space. For example, 255 is encoded as

	nBytes 4 0 0 0 255

This representation is not as space-efficient as others, but improving its space
usage didn't seem worth the additional complexity.

Signed integers use gob's encoding.

Floats are encoded by converting them to uints using math.Float64bits.

The nValues code is for sequences of values whose size is known beforehand,
like a Go slice or array. The slice []string{"hi", "bye"} is encoded as

	nValues 2 nBytes 2 'h' 'i' nBytes 3 'b' 'y' 'e'

The ref code is used to refer to an earlier encoded value. It is followed by
a uint denoting the index data of the value to use.

The start and end codes delimit a value whose length is unknown beforehand.
They are used for structs.

A struct is encoded as a sequence of fields between a start code and an end code.
Each exported field is encoded as an integer field number followed by the field's value.
Fields with zero values are omitted.

The field number is initially the position of the field in the struct declaration. This
initial ordering is preserved even if fields are added or removed; new fields are
numbered after the initial ones. This is accomplished by preserving the field ordering
in a comment in the generated code. For example, say a struct S has fields A and B.
The generated encoder would assign 0 to A and 1 to B, and include this comment in the code:

	// Fields of S: A B

If later a field X was added between A and B, the generator would read the comment
and preserve the original assignments of A and B. It would assign 2 to X, and the
generated code would contain the new comment

	// Fields of S: A B X

Values of type any are encoded as a pair of an assigned type number and the value.
All types that can appear as values must be registered by calling Register.
Type numbers are assigned to type names by the encoder in the order that types are encountered.
The assignments are saved at the beginning of the encoded data.
*/
package codec
