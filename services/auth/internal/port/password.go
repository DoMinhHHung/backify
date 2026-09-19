package port

// PasswordHasher bọc hash/verify mật khẩu (argon2id). Package password
// cung cấp Hasher{} với method gọi thẳng hàm cấp package Hash/Verify vì
// hàm cấp package không tự thỏa interface được.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (bool, error)
}
