# 🤖 Panduan Lengkap Setup Discord Bot Minecraft dari Nol (0)

Dokumen ini akan memandu Anda langkah demi langkah tentang cara mengatur dan menjalankan Bot Discord Minecraft Anda, khususnya cara mendapatkan ID dan Token ajaib yang dibutuhkan di dalam file `.env`.

---

## 📁 Struktur File `.env`
File `.env` adalah nyawa dari bot Anda. File ini menyimpan kata sandi rahasia agar bot bisa hidup dan nyambung ke server Minecraft. Bentuknya seperti ini:

```env
BOT_TOKEN=MTEyMz... (Rahasia)
SERVICE_NAME=minecraft
GUILD_ID=1193478717947793418
RCON_PASSWORD=RahasiaMinecraft123
SYNC_CHANNEL_ID=1539703779451076720
MINECRAFT_PATH=//home/minecraft/server-minecraft
OWNER_ID=755261155903012925
```

Berikut adalah cara mendapatkan masing-masing variabel tersebut:

---

## 1. `BOT_TOKEN` (Kunci Nyawa Bot)
Ini adalah tiket masuk agar program Golang kita bisa mengontrol Bot di Discord.
**Cara Mendapatkan:**
1. Buka [Discord Developer Portal](https://discord.com/developers/applications).
2. Login menggunakan akun Discord Anda, lalu klik tombol **New Application** di sudut kanan atas. Beri nama bot Anda (misal: *Minecraft Penjaga*).
3. Di menu sebelah kiri, klik tab **Bot**.
4. Di halaman Bot, Anda akan melihat tombol **Reset Token**. Klik tombol itu, lalu klik **Yes, do it!**.
5. Sebuah kode panjang (Token) akan muncul. **Copy** kode tersebut dan *paste* ke `BOT_TOKEN` di file `.env`.
> [!WARNING]
> Jangan pernah membagikan `BOT_TOKEN` Anda kepada siapa pun! Jika ada yang tahu token ini, mereka bisa menghack bot Anda.

## 2. `GUILD_ID` (ID Server Discord)
Bot perlu tahu di Server Discord mana dia sedang bekerja agar perintah *Slash Command* (`/gelar`, `/whereis`) hanya muncul di server Anda.
**Cara Mendapatkan:**
1. Buka aplikasi Discord Anda.
2. Pergi ke **User Settings** (ikon roda gigi) > **Advanced** > Nyalakan **Developer Mode**.
3. Tutup pengaturan, lalu pergi ke Server Discord komunitas Anda.
4. Klik kanan pada **Nama Server** Anda (di pojok kiri atas).
5. Klik **Copy Server ID**. Paste angka tersebut ke `GUILD_ID`.

## 3. `OWNER_ID` (ID Anda Sebagai Bos)
Ini untuk fitur keamanan, agar hanya Anda (Sang Admin) yang bisa mengeksekusi perintah sensitif seperti `/whereis` atau `/tourney`.
**Cara Mendapatkan:**
1. Pastikan **Developer Mode** di Discord Anda sudah menyala (seperti langkah di atas).
2. Kirim pesan apa saja di chat Discord.
3. Klik kanan pada **Foto Profil (Avatar) Anda** sendiri di pesan tersebut.
4. Klik **Copy User ID**. Paste angka tersebut ke `OWNER_ID`.

## 4. `SYNC_CHANNEL_ID` (Jalur Lintas Chat)
Ini adalah ID ruangan (Channel) di Discord tempat bot mengirimkan chat dari pemain Minecraft, dan sebaliknya (Cross-Chat).
**Cara Mendapatkan:**
1. Buat atau pilih satu *Text Channel* khusus di Discord Anda (misal: `#mc-chat`).
2. Klik kanan pada nama channel tersebut.
3. Klik **Copy Channel ID**. Paste ke `SYNC_CHANNEL_ID`.

## 5. `RCON_PASSWORD` (Kunci Remote Minecraft)
Agar bot bisa mengirim perintah `/team`, `/tp`, atau `/give` ke dalam game, kita menggunakan fitur **RCON** bawaan Minecraft.
**Cara Mendapatkan & Mengatur:**
1. Buka file `server.properties` yang ada di dalam folder server Minecraft Anda.
2. Cari baris `enable-rcon=false` lalu ubah menjadi `enable-rcon=true`.
3. Cari baris `rcon.password=` lalu isi dengan sandi rahasia Anda (bebas, misal: `rcon.password=RahasiaMinecraft123`).
4. Samakan sandi tersebut ke dalam file `.env` di bagian `RCON_PASSWORD`.
5. Restart server Minecraft Anda agar RCON aktif.

## 6. `MINECRAFT_PATH` & `SERVICE_NAME`
- `SERVICE_NAME`: Jika menggunakan OS Linux dan menjalankan server lewat systemctl, biarkan nilainya `minecraft` atau sesuaikan dengan nama *service* Anda (misal `minecraft-server`). Ini digunakan oleh bot untuk membaca log kematian (`journalctl -u minecraft`).
- `MINECRAFT_PATH`: Ini adalah alamat folder (direktori) lengkap tempat server Minecraft Anda bersarang (tempat di mana ada file `server.properties`, `world/`, dll). Bot membutuhkan akses ini untuk membaca file *Stats* dan *Usercache* untuk membagikan gelar secara otomatis. (Misal: `/home/minecraft/server-minecraft`).

---

## 🚀 Cara Mengaktifkan (Menyalakan) Bot di VPS

Jika semua data di dalam `.env` sudah terisi sempurna, berikut adalah urutan menyalakan bot-nya:

1. Buka Terminal/SSH VPS Anda.
2. Masuk ke folder tempat kode bot ini berada: `cd /path/ke/folder/mcbot`
3. Ketik perintah ini untuk memastikan semua *library* pendukung Go diunduh:
   ```bash
   go mod tidy
   ```
4. *Compile* (rakit) kode bot menjadi satu program jadi:
   ```bash
   go build -o botku .
   ```
5. Nyalakan bot-nya!
   ```bash
   ./botku
   ```
Jika muncul pesan `[BOT] Siap melayani tuan...` di layar hitam tersebut, berarti bot Anda telah hidup dan siap mendominasi server! 🎉

---
## 7. Cara Mengundang Bot ke Server & Setting Permissions
Saat Anda membuat link undangan untuk memasukkan bot ke server Discord Anda, pastikan Anda mencentang izin (Permissions) berikut di tab **OAuth2 -> URL Generator**:

**Di bagian SCOPES:**
- [x] ot
- [x] pplications.commands (Wajib agar Slash Command seperti /gelar bisa muncul)

**Di bagian BOT PERMISSIONS:**
- [x] View Channels (Read Messages)
- [x] Send Messages
- [x] Embed Links (Opsional untuk tampilan keren)
- [x] Read Message History

**⚠️ SANGAT PENTING (Di Tab BOT):**
Pergi ke menu **Bot** di sebelah kiri (di bawah OAuth2), lalu *scroll* ke bawah sampai Anda menemukan bagian **Privileged Gateway Intents**.
Pastikan Anda menyalakan (mencentang biru) opsi:
- ✅ **MESSAGE CONTENT INTENT**
*(Jika ini dimatikan, fitur Cross-Chat tidak akan berfungsi karena bot Anda menjadi buta dan tidak bisa membaca chat pemain dari Discord).*
