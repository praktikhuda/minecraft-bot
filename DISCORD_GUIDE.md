# 🎮 Minecraft Discord Bot Guide

Bot Discord ini dirancang sebagai "Remote Control" jarak jauh dan pusat informasi untuk server Minecraft Anda. Bot ini mengontrol server secara langsung menggunakan integrasi `mcrcon`, `systemctl`, serta analisis *file* lokal.

---

## 📋 Daftar Fitur Lengkap

### ⚙️ Manajemen Server (Kontrol Penuh VPS)
1. **`/start`** - Menyalakan server Minecraft (*systemctl start*).
2. **`/stop`** - Mematikan server Minecraft dengan aman (*systemctl stop*).
3. **`/restart`** - Memuat ulang server (*systemctl restart*).
4. **`/status`** - Mengecek apakah server sedang aktif atau mati.
5. **`/info`** - Mengecek persentase penggunaan **CPU dan RAM** mesin VPS secara *real-time*.
6. **`/logs`** - Membaca *log* internal Minecraft langsung dari Discord tanpa perlu SSH (`journalctl`).
7. **`/backup`** - Membuat cadangan (*zip*) folder *world* Minecraft otomatis.


### 🛡️ Moderasi & Interaksi Pemain
10. **`/kick <pemain> <alasan>`** - Mengeluarkan pemain yang sedang *online*.
11. **`/ban <pemain> <alasan>`** - Mencekal pemain (Blokir permanen).
12. **`/unban <pemain>`** - Membuka cekal pemain.
13. **`/wl <add/remove/list> <pemain>`** - Mengelola *Whitelist* (Daftar pemain eksklusif).
14. **`/say <pesan>`** - Mengirim pesan resmi (Broadcast) ke dalam *game* sebagai **[Server]**.
15. **`/op <pemain> <role>`** - **(Khusus Owner)** Mengatur jabatan pemain. Pilihan:
    - **Admin** (Dewa: OP + Creative)
    - **Spectator** (Mata-mata: Deop + Spectator)
    - **Player** (Normal: Deop + Survival)
16. **`/deop <pemain>`** - **(Khusus Owner)** Mencabut hak akses OP secara cepat.
17. **Sistem Cross-Chat** - Komunikasi tanpa batas. Pesan di channel Discord dikirim ke dalam game, dan obrolan *in-game* dikirim kembali ke Discord!

### 🏆 Sistem Statistik & Gelar (Leaderboard & LuckPerms)
18. **`/leaderboard`** - Mengekstrak *file* `stats` pemain dan menyajikan Top Papan Peringkat:
    - Kematian Terbanyak (*Deaths*)
    - Pembunuh Terbanyak (*Player Kills*)
    - Waktu Bermain (*Playtime*)
    - Diamond Tergali (*Diamonds Mined*)
    - Penggunaan Totem (*Totems Popped*)
19. **`/sync_titles`** - Secara otomatis mengecek `/leaderboard` dan memberikan gelar/pangkat (*Prefix*) di dalam game menggunakan **LuckPerms** (Contoh: `[Pencabut Nyawa]`, `[Sultan]`).

### 🎪 Sistem Event Spesial
20. **`/blackmarket`** - Memanggil NPC kustom *"Pedagang Gelap"* di koordinat `X:-10 Y:75 Z:5`. NPC ini lumpuh (`movement_speed: 0.0`) dan kebal, menjual *Elytra* seharga 1 Kepala Wither + 64 Diamond.
21. **`/rmblackmarket`** - Membersihkan seluruh NPC *"Pedagang Gelap"* yang ada di dunia dengan aman.

---

> **Catatan Keamanan:** Perintah sensitif seperti `/op` dan `/deop` dikunci rapat-rapat. Meskipun ada Admin Discord lain yang mencoba menggunakannya, bot akan menolak kecuali ID Discord Anda (Sebagai *Owner*) yang menekannya.
