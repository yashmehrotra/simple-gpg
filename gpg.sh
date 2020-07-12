# To encrypt
gpg --symmetric --cipher-algo AES256 $1

# To decrypt
gpg $1
