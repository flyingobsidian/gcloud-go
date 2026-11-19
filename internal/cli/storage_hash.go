package cli

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"google.golang.org/api/googleapi"
	storage "google.golang.org/api/storage/v1"
)

// --- Storage hash validation (#1943) ---
//
// gcloud-python 589.0.0 sends upload hashes for server-side validation and
// checks composed objects' CRC32C against their sources.

// castagnoliReversed is the reflected CRC32C polynomial.
const castagnoliReversed = 0x82F63B78

// storageBoolProperty reads storage/<key> as a gcloud boolean property,
// returning def when it is unset.
func (s *storageCmds) storageBoolProperty(key string, def bool) (bool, error) {
	v, err := s.inv.env.Property("storage", key)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "":
		return def, nil
	case "1", "true", "on", "yes", "y":
		return true, nil
	case "0", "false", "off", "no", "n":
		return false, nil
	}
	return false, fmt.Errorf("the [storage/%s] value [%s] is not valid. Possible values: [true, 1, on, yes, y, false, 0, off, no, n, '', none]. (See http://yaml.org/type/bool.html)", key, v)
}

// storageCheckHashesNever reports whether storage/check_hashes is "never".
func (s *storageCmds) storageCheckHashesNever() (bool, error) {
	v, err := s.inv.env.Property("storage", "check_hashes")
	if err != nil {
		return false, err
	}
	return v == "never", nil
}

// gf2MatrixTimes multiplies the GF(2) matrix mat by vec.
func gf2MatrixTimes(mat *[32]uint32, vec uint32) uint32 {
	var sum uint32
	for i := 0; vec != 0; i, vec = i+1, vec>>1 {
		if vec&1 != 0 {
			sum ^= mat[i]
		}
	}
	return sum
}

// gf2MatrixSquare sets square to mat*mat.
func gf2MatrixSquare(square, mat *[32]uint32) {
	for n := range 32 {
		square[n] = gf2MatrixTimes(mat, mat[n])
	}
}

// crc32cCombine returns the CRC32C of A||B given crc1 = CRC32C(A),
// crc2 = CRC32C(B) and len2 = len(B), as zlib's crc32_combine does.
func crc32cCombine(crc1, crc2 uint32, len2 int64) uint32 {
	if len2 <= 0 {
		return crc1
	}
	var even, odd [32]uint32
	odd[0] = castagnoliReversed
	row := uint32(1)
	for n := 1; n < 32; n++ {
		odd[n] = row
		row <<= 1
	}
	gf2MatrixSquare(&even, &odd) // two zero bits
	gf2MatrixSquare(&odd, &even) // four zero bits
	for {
		gf2MatrixSquare(&even, &odd)
		if len2&1 != 0 {
			crc1 = gf2MatrixTimes(&even, crc1)
		}
		len2 >>= 1
		if len2 == 0 {
			break
		}
		gf2MatrixSquare(&odd, &even)
		if len2&1 != 0 {
			crc1 = gf2MatrixTimes(&odd, crc1)
		}
		len2 >>= 1
		if len2 == 0 {
			break
		}
	}
	return crc1 ^ crc2
}

// decodeCRC32C decodes a base64 big-endian CRC32C as returned by the API.
func decodeCRC32C(b64 string) (uint32, error) {
	b, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return 0, err
	}
	if len(b) != 4 {
		return 0, fmt.Errorf("invalid CRC32C length %d", len(b))
	}
	return binary.BigEndian.Uint32(b), nil
}

// encodeCRC32C encodes a CRC32C as base64 big-endian.
func encodeCRC32C(crc uint32) string {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], crc)
	return base64.StdEncoding.EncodeToString(b[:])
}

// expectedComposedCRC32C returns the CRC32C a composition of sources should
// have, or ok=false (after warning) when a source lacks a CRC32C or size,
// as validate_composed_hash_matches does.
func (s *storageCmds) expectedComposedCRC32C(sources []*storage.Object) (string, bool, error) {
	if len(sources) == 0 {
		fmt.Fprintln(s.inv.stderr, "WARNING: Could not perform checksum validation for composed object due to missing destination or source resources.")
		return "", false, nil
	}
	var crc uint32
	for i, src := range sources {
		if src.Crc32c == "" {
			fmt.Fprintln(s.inv.stderr, "WARNING: Could not perform checksum validation for composed object due to missing metadata (crc32c_hash or size) in source.")
			return "", false, nil
		}
		next, err := decodeCRC32C(src.Crc32c)
		if err != nil {
			return "", false, fmt.Errorf("Failed to calculate expected checksum: %w", err)
		}
		if i == 0 {
			crc = next
			continue
		}
		crc = crc32cCombine(crc, next, int64(src.Size))
	}
	return encodeCRC32C(crc), true, nil
}

// validateComposedHash checks the composed object's CRC32C against its
// sources unless storage/check_hashes is never.
func (s *storageCmds) validateComposedHash(sources []*storage.Object, composed *storage.Object) error {
	never, err := s.storageCheckHashesNever()
	if err != nil || never {
		return err
	}
	if composed == nil {
		fmt.Fprintln(s.inv.stderr, "WARNING: Could not perform checksum validation for composed object due to missing destination or source resources.")
		return nil
	}
	expected, ok, err := s.expectedComposedCRC32C(sources)
	if err != nil || !ok {
		return err
	}
	if composed.Crc32c == "" {
		fmt.Fprintln(s.inv.stderr, "WARNING: Skipping checksum validation for composed object due to missing metadata (crc32c_hash) in destination.")
		return nil
	}
	if expected != composed.Crc32c {
		return fmt.Errorf("Source hash %s does not match destination hash %s for object gs://%s/%s.",
			expected, composed.Crc32c, composed.Bucket, composed.Name)
	}
	return nil
}

// usesResumableUpload reports whether the Go client uploads size bytes with
// a resumable session (more than one chunk) rather than a single request.
func usesResumableUpload(size int64) bool {
	return size > googleapi.DefaultUploadChunkSize
}

// prepareServerSideHashValidation sets obj.Md5Hash so the server validates a
// resumable upload of f, and returns a reader that fails if f's size changes,
// as gcloud-python's final_headers_callback does. It returns f unchanged when
// storage/enable_server_side_hash_validation is false, storage/check_hashes
// is never, or the upload is not resumable.
func (s *storageCmds) prepareServerSideHashValidation(f *os.File, size int64, obj *storage.Object) (io.Reader, error) {
	if !usesResumableUpload(size) {
		return f, nil
	}
	enabled, err := s.storageBoolProperty("enable_server_side_hash_validation", true)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return f, nil
	}
	never, err := s.storageCheckHashesNever()
	if err != nil {
		return nil, err
	}
	reader := &sizeCheckReader{f: f, size: size}
	if never {
		return reader, nil
	}
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("hashing local file: %w", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewinding local file: %w", err)
	}
	obj.Md5Hash = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return reader, nil
}

// sizeCheckReader reads f and, at EOF, fails if f's size on disk no longer
// matches size.
type sizeCheckReader struct {
	f    *os.File
	size int64
}

func (r *sizeCheckReader) Read(p []byte) (int, error) {
	n, err := r.f.Read(p)
	if errors.Is(err, io.EOF) {
		if info, statErr := os.Stat(r.f.Name()); statErr == nil && info.Mode().IsRegular() && info.Size() != r.size {
			return n, errors.New("Source file size changed post-initiation.")
		}
	}
	return n, err
}
