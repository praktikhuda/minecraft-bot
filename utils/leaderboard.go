package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"os/exec"
	"log"
)

type LBCategory struct {
	Name    string
	StatID  string
	Desc    string
	Title   string
	Stars   int
}

var (
	LBCombat  = []LBCategory{}
	LBHunt    = []LBCategory{}
	LBMining  = []LBCategory{}
	LBExplore = []LBCategory{}
	LBFood    = []LBCategory{}
	LBFails   = []LBCategory{}
	LBMisc    = []LBCategory{}
	LBExtra   = []LBCategory{}
)

func init() {
	LBCombat = append(LBCombat, LBCategory{"Raja PVP", "minecraft:custom:player_kills", "Paling banyak membunuh pemain lain.", "Raja PVP", 2})
	LBCombat = append(LBCombat, LBCategory{"Ahli Panah", "minecraft:used:bow", "Paling sering menembakkan panah.", "Robin Hood", 2})
	LBCombat = append(LBCombat, LBCategory{"Sniper Jarak Jauh", "minecraft:custom:target_hit", "Mengenai target dari jarak paling jauh.", "Sniper", 2})
	LBCombat = append(LBCombat, LBCategory{"Penguasa Trident", "minecraft:used:trident", "Paling sering melempar Trident.", "Aquaman", 2})
	LBCombat = append(LBCombat, LBCategory{"Sembilan Nyawa", "minecraft:used:totem_of_undying", "Paling sering selamat menggunakan Totem.", "Sembilan Nyawa", 4})
	LBCombat = append(LBCombat, LBCategory{"Si Paling Samsak", "minecraft:custom:damage_taken", "Menerima kerusakan total paling besar (Damage Taken).", "Samsak Hidup", 2})
	LBCombat = append(LBCombat, LBCategory{"Haus Darah", "minecraft:custom:damage_dealt", "Memberikan kerusakan total paling besar ke entitas apa saja.", "Haus Darah", 2})
	LBCombat = append(LBCombat, LBCategory{"Ahli Pedang", "minecraft:used:netherite_sword", "Paling sering menggunakan pedang Netherite.", "Ahli Pedang", 4})
	LBCombat = append(LBCombat, LBCategory{"Si Paling Rapuh", "minecraft:custom:damage_resisted", "Paling sedikit menahan damage sebelum mati.", "Kerupuk", 2})
	LBHunt = append(LBHunt, LBCategory{"Pembantai Zombie", "minecraft:killed:zombie", "Paling banyak membunuh Zombie.", "Zombie Hunter", 2})
	LBHunt = append(LBHunt, LBCategory{"Pembantai Skeleton", "minecraft:killed:skeleton", "Paling banyak mematahkan tulang Skeleton.", "Patah Tulang", 2})
	LBHunt = append(LBHunt, LBCategory{"Pembantai Creeper", "minecraft:killed:creeper", "Penyelamat server dari Creeper.", "Penjinak Bom", 2})
	LBHunt = append(LBHunt, LBCategory{"Pembunuh Naga", "minecraft:killed:ender_dragon", "Paling banyak memberikan serangan mematikan ke Ender Dragon.", "Dragon Slayer", 5})
	LBHunt = append(LBHunt, LBCategory{"Penjaga Desa", "minecraft:killed:pillager", "Paling banyak membunuh gerombolan Pillager.", "Pahlawan Desa", 2})
	LBHunt = append(LBHunt, LBCategory{"Gladiator", "minecraft:killed:wither", "Paling sering membunuh boss Wither.", "Penakluk Wither", 5})
	LBHunt = append(LBHunt, LBCategory{"Pemburu Phantom", "minecraft:killed:phantom", "Paling banyak membunuh Phantom karena jarang tidur.", "Satpam Malam", 3})
	LBHunt = append(LBHunt, LBCategory{"Tukang Tipu Piglin", "minecraft:killed:piglin", "Paling banyak membunuh Piglin di Nether.", "Musuh Nether", 2})
	LBHunt = append(LBHunt, LBCategory{"Pemburu Enderman", "minecraft:killed:enderman", "Paling banyak membunuh Enderman untuk Ender Pearl.", "Mata Merah", 2})
	LBMining = append(LBMining, LBCategory{"Juragan Diamond", "minecraft:mined:diamond_ore", "Menambang Diamond Ore terbanyak.", "Juragan Diamond", 4})
	LBMining = append(LBMining, LBCategory{"Pencari Netherite", "minecraft:mined:ancient_debris", "Menambang Ancient Debris terbanyak.", "Kuli Nether", 5})
	LBMining = append(LBMining, LBCategory{"Kuli Batu", "minecraft:mined:stone", "Menambang batu (Cobblestone/Stone) terbanyak.", "Kuli Bangunan", 2})
	LBMining = append(LBMining, LBCategory{"Tukang Gali", "minecraft:mined:dirt", "Menggali tanah (Dirt) terbanyak.", "Cacing Tanah", 1})
	LBMining = append(LBMining, LBCategory{"Raja Hutan", "minecraft:mined:oak_log", "Menebang kayu paling banyak.", "Penebang Pohon", 2})
	LBMining = append(LBMining, LBCategory{"Penambang Emas", "minecraft:mined:gold_ore", "Paling banyak menambang emas.", "Sultan Emas", 3})
	LBMining = append(LBMining, LBCategory{"Petani Gandum", "minecraft:mined:wheat", "Paling banyak memanen gandum.", "Petani Desa", 2})
	LBMining = append(LBMining, LBCategory{"Petani Wortel", "minecraft:mined:carrots", "Paling banyak memanen wortel.", "Si Kelinci", 2})
	LBMining = append(LBMining, LBCategory{"Pemecah Pickaxe", "minecraft:broken:diamond_pickaxe", "Paling sering merusakkan beliung diamond/netherite.", "Penghancur Alat", 4})
	LBMining = append(LBMining, LBCategory{"Penggila Redstone", "minecraft:mined:redstone_ore", "Menambang Redstone terbanyak.", "Ahli Mesin", 2})
	LBMining = append(LBMining, LBCategory{"Tukang Gali Pasir", "minecraft:mined:sand", "Menambang pasir paling banyak untuk membuat kaca/TNT.", "Penyapu Gurun", 1})
	LBMining = append(LBMining, LBCategory{"Penambang Lapis", "minecraft:mined:lapis_ore", "Menambang Lapis Lazuli terbanyak.", "Si Penyihir", 2})
	LBMining = append(LBMining, LBCategory{"Pembersih Nether", "minecraft:mined:netherrack", "Paling banyak menggali Netherrack.", "Penjelajah Neraka", 2})
	LBExplore = append(LBExplore, LBCategory{"Tukang Kabur", "minecraft:custom:leave_game", "Siapa yang paling sering logout / keluar-masuk server.", "Tukang Kabur", 2})
	LBExplore = append(LBExplore, LBCategory{"Si Paling Panik", "minecraft:custom:crouch_one_cm", "Paling sering jalan jongkok (Sneaking) karena ketakutan.", "Ninja Kesasar", 2})
	LBExplore = append(LBExplore, LBCategory{"Phobia Ketinggian", "minecraft:custom:climb_one_cm", "Jarak terjauh memanjat tangga atau vine.", "Tukang Panjat", 2})
	LBExplore = append(LBExplore, LBCategory{"Hobi Cuci Muka", "minecraft:custom:walk_under_water_one_cm", "Jarak berjalan di dasar laut.", "SpongeBob", 2})
	LBExplore = append(LBExplore, LBCategory{"Si Paling Sibuk", "minecraft:custom:walk_one_cm", "Jarak berjalan kaki paling jauh.", "Pejalan Kaki", 2})
	LBExplore = append(LBExplore, LBCategory{"Pelari Cepat", "minecraft:custom:sprint_one_cm", "Jarak berlari (sprinting) paling jauh.", "Pelari Maraton", 2})
	LBExplore = append(LBExplore, LBCategory{"Pengelana Air", "minecraft:custom:boat_one_cm", "Jarak mendayung perahu paling jauh.", "Pelaut", 2})
	LBExplore = append(LBExplore, LBCategory{"Penunggang Babi", "minecraft:custom:pig_one_cm", "Jarak menunggang babi paling jauh pakai Carrot on a Stick.", "Penunggang Babi", 2})
	LBExplore = append(LBExplore, LBCategory{"Joki Kuda", "minecraft:custom:horse_one_cm", "Jarak menunggang kuda paling jauh.", "Koboi Server", 2})
	LBExplore = append(LBExplore, LBCategory{"Pilot Handal", "minecraft:custom:aviate_one_cm", "Jarak terbang pakai Elytra paling jauh.", "Manusia Burung", 2})
	LBExplore = append(LBExplore, LBCategory{"Perenang Cepat", "minecraft:custom:swim_one_cm", "Jarak berenang paling jauh.", "Atlet Renang", 2})
	LBExplore = append(LBExplore, LBCategory{"Tukang Lompat Es", "minecraft:custom:walk_on_water_one_cm", "Berjalan/melompat di atas air (Frost Walker).", "Manusia Es", 2})
	LBExplore = append(LBExplore, LBCategory{"Pengembara Nether", "minecraft:custom:strider_one_cm", "Menunggangi Strider di lautan Lava.", "Penunggang Lava", 3})
	LBExplore = append(LBExplore, LBCategory{"Tukang Nyasar", "minecraft:custom:play_time", "Total waktu bermain (*Play Time*) paling tinggi.", "Sepuh Server", 5})
	LBExplore = append(LBExplore, LBCategory{"Turis Dimensi", "minecraft:custom:time_since_last_rest", "Waktu paling lama bertahan hidup tanpa tidur.", "Mata Panda", 2})
	LBFood = append(LBFood, LBCategory{"Makan Sembarangan", "minecraft:used:rotten_flesh", "Paling banyak memakan daging busuk (Rotten Flesh).", "Perut Besi", 2})
	LBFood = append(LBFood, LBCategory{"Tukang Makan", "minecraft:custom:eat_record_count", "Mengkonsumsi makanan paling banyak.", "Tukang Makan", 2})
	LBFood = append(LBFood, LBCategory{"Pemakan Apel Emas", "minecraft:used:golden_apple", "Paling banyak mengkonsumsi Golden Apple.", "Gigi Emas", 3})
	LBFood = append(LBFood, LBCategory{"Pemakan Apel Enchant", "minecraft:used:enchanted_golden_apple", "Paling banyak makan Notch Apple.", "Sultan Apel", 4})
	LBFood = append(LBFood, LBCategory{"Tukang Roti", "minecraft:used:bread", "Paling sering memakan Roti.", "Pecinta Karbo", 2})
	LBFood = append(LBFood, LBCategory{"Pemabuk Potion", "minecraft:used:potion", "Paling banyak meminum ramuan (Potion).", "Dukun", 2})
	LBFood = append(LBFood, LBCategory{"Pecandu Madu", "minecraft:used:honey_bottle", "Paling banyak meminum madu.", "Winnie the Pooh", 2})
	LBFails = append(LBFails, LBCategory{"Korban Ledakan", "minecraft:killed_by:creeper", "Siapa yang paling sering mati karena pelukan hangat Creeper.", "Pecinta Creeper", 2})
	LBFails = append(LBFails, LBCategory{"Jatuh Dari Langit", "minecraft:custom:fall_one_cm", "Jarak terjauh (akumulasi) saat jatuh.", "Penerjun Bebas", 2})
	LBFails = append(LBFails, LBCategory{"Mandi Lava", "minecraft:killed_by:lava", "Paling sering mati terbakar karena berenang di Lava.", "Tahan Panas", 2})
	LBFails = append(LBFails, LBCategory{"Tersambar Petir", "minecraft:killed_by:lightning_bolt", "Pemain paling sial yang sering disambar petir.", "Anak Zeus", 2})
	LBFails = append(LBFails, LBCategory{"Dilempar Kinetik", "minecraft:killed_by:fly_into_wall", "Paling sering mati karena menabrak dinding saat pakai Elytra.", "Pilot Mabuk", 2})
	LBFails = append(LBFails, LBCategory{"Korban Llama", "minecraft:killed_by:llama_spit", "Paling sering diludahi Llama sampai mati.", "Diludahi", 2})
	LBFails = append(LBFails, LBCategory{"Si Paling Tumbal", "minecraft:custom:deaths", "Paling banyak mati di tangan pemain lain atau mob.", "Si Paling Tumbal", 2})
	LBFails = append(LBFails, LBCategory{"Korban Gravitasi", "minecraft:killed_by:fall", "Mati karena jatuh paling sering.", "Tidak Punya Parasut", 2})
	LBFails = append(LBFails, LBCategory{"Lemah Jantung", "minecraft:killed_by:wither_skeleton", "Paling sering dibunuh Wither Skeleton.", "Kena Mental", 5})
	LBFails = append(LBFails, LBCategory{"Korban Kaktus", "minecraft:killed_by:cactus", "Terbunuh kaktus. Ya, Kaktus.", "Tertusuk Kaktus", 2})
	LBFails = append(LBFails, LBCategory{"Mati Konyol", "minecraft:killed_by:cramming", "Mati karena tergencet di tempat sempit dengan banyak entity.", "Sarden", 2})
	LBFails = append(LBFails, LBCategory{"Mati Tertimpa Anvil", "minecraft:killed_by:falling_block", "Tertimpa balok jatuh (Pasir / Gravel / Anvil).", "Kepala Batu", 2})
	LBFails = append(LBFails, LBCategory{"Si Buta", "minecraft:custom:time_since_death", "Paling cepat mati setelah respawn.", "Speedrun Mati", 2})
	LBMisc = append(LBMisc, LBCategory{"Si Paling Capek", "minecraft:custom:jump", "Siapa yang paling banyak melompat di dalam server.", "Si Kutu Loncat", 2})
	LBMisc = append(LBMisc, LBCategory{"Kaum Rebahan", "minecraft:custom:sleep_in_bed", "Siapa yang paling rajin tidur di kasur untuk skip malam.", "Kaum Rebahan", 2})
	LBMisc = append(LBMisc, LBCategory{"Cinta Damai", "minecraft:custom:animals_bred", "Paling sering mengawinkan hewan.", "Peternak Cinta", 2})
	LBMisc = append(LBMisc, LBCategory{"Tukang Mancing", "minecraft:custom:fish_caught", "Paling banyak menangkap ikan.", "Pemancing Handal", 2})
	LBMisc = append(LBMisc, LBCategory{"Pencari Harta", "minecraft:custom:inspect_hopper", "Paling sering membuka Hopper / Chest (Looting).", "Tukang Intip", 2})
	LBMisc = append(LBMisc, LBCategory{"Juragan Kampung", "minecraft:custom:traded_with_villager", "Paling sering jual-beli (Trade) dengan Villager.", "Milyarder Desa", 2})
	LBMisc = append(LBMisc, LBCategory{"Pembakar Makanan", "minecraft:used:furnace", "Paling banyak memasak di Furnace.", "Koki Server", 2})
	LBMisc = append(LBMisc, LBCategory{"Pencari XP", "minecraft:custom:total_world_time", "Paling banyak mendapatkan Experience Orb.", "Level Dewa", 2})
	LBMisc = append(LBMisc, LBCategory{"Si Paling Penempatan", "minecraft:used:cobblestone", "Paling banyak meletakkan blok Cobblestone.", "Tukang Batu", 1})
	LBMisc = append(LBMisc, LBCategory{"Desainer Interior", "minecraft:used:oak_planks", "Paling banyak menggunakan blok papan kayu (Planks).", "Tukang Kayu", 2})
	LBMisc = append(LBMisc, LBCategory{"Tukang Kaca", "minecraft:used:glass", "Paling banyak memasang blok kaca.", "Pemasang Jendela", 2})
	LBMisc = append(LBMisc, LBCategory{"Pembakar Obor", "minecraft:used:torch", "Paling banyak menerangi dunia dengan memasang Torch.", "Pembawa Cahaya", 2})
	LBMisc = append(LBMisc, LBCategory{"Ahli Beton", "minecraft:used:cyan_concrete", "Paling banyak menggunakan blok Concrete untuk membangun modern.", "Arsitek Modern", 2})
	LBMisc = append(LBMisc, LBCategory{"Pembuat Pintu", "minecraft:used:iron_door", "Paling banyak memasang Iron Door.", "Si Paling Tertutup", 2})
	LBMisc = append(LBMisc, LBCategory{"Tukang Ledak", "minecraft:used:tnt", "Paling banyak menyalakan blok TNT.", "Teroris Server", 2})
	LBMisc = append(LBMisc, LBCategory{"Pemasang Kasur", "minecraft:used:red_bed", "Paling banyak menaruh kasur.", "Juragan Kos", 2})
	LBMisc = append(LBMisc, LBCategory{"Kolektor Musik", "minecraft:used:music_disc_13", "Paling sering memutar kaset Music Disc di Jukebox.", "DJ Server", 2})
	LBMisc = append(LBMisc, LBCategory{"Petasan Mania", "minecraft:used:firework_rocket", "Paling sering menyalakan kembang api (termasuk untuk Elytra).", "Tukang Petasan", 2})
	LBMisc = append(LBMisc, LBCategory{"Ahli Lonceng", "minecraft:custom:bell_ring", "Paling sering membunyikan lonceng desa.", "Penjaga Sekolah", 2})
	LBMisc = append(LBMisc, LBCategory{"Seniman Banner", "minecraft:used:loom", "Paling sering memakai mesin Loom untuk membuat Banner.", "Seniman", 2})
	LBExtra = append(LBExtra, LBCategory{"Pembuat Peta", "minecraft:used:cartography_table", "Paling sering menggunakan meja peta.", "Sang Kartografer", 2})
	LBExtra = append(LBExtra, LBCategory{"Pandai Besi", "minecraft:used:anvil", "Paling banyak menghabiskan level di Anvil untuk memperbaiki alat.", "Pandai Besi", 2})
	LBExtra = append(LBExtra, LBCategory{"Tukang Enchanter", "minecraft:used:enchanting_table", "Paling sering melakukan Enchantment di meja sihir.", "Sang Penyihir", 2})
	LBExtra = append(LBExtra, LBCategory{"Penebar Bunga", "minecraft:used:poppy", "Paling banyak menanam bunga.", "Tukang Kebun", 2})
	LBExtra = append(LBExtra, LBCategory{"Pewarna Domba", "minecraft:used:red_dye", "Paling banyak mewarnai bulu domba.", "Tukang Cat", 2})
	LBExtra = append(LBExtra, LBCategory{"Pengepul Shulker", "minecraft:used:shulker_box", "Paling sering buka-tutup Shulker Box.", "Pengepul Barang", 2})
	LBExtra = append(LBExtra, LBCategory{"Tukang Nge-Camp", "minecraft:used:campfire", "Paling banyak meletakkan Campfire.", "Anak Pramuka", 2})
	LBExtra = append(LBExtra, LBCategory{"Raja Sampah", "minecraft:dropped:cobblestone", "Paling sering membuang item sembarangan (Drop Item).", "Pembuang Sampah", 1})
	LBExtra = append(LBExtra, LBCategory{"Tukang Bersih-Bersih", "minecraft:picked_up:cobblestone", "Memungut block paling banyak.", "Pemulung", 1})
	LBExtra = append(LBExtra, LBCategory{"Penggila Telur", "minecraft:used:egg", "Paling sering melempar telur ayam.", "Tukang Lempar", 2})
	LBExtra = append(LBExtra, LBCategory{"Pengendali Salju", "minecraft:used:snowball", "Paling banyak melempar bola salju.", "Elsa", 2})
	LBExtra = append(LBExtra, LBCategory{"Penjaga Ender Chest", "minecraft:used:ender_chest", "Paling sering buka Ender Chest karena paranoid kehilangan barang.", "Paranoid", 2})

}

type PlayerStat struct {
	UUID  string
	Name  string
	Value int
}

func GetAllCategories() []LBCategory {
	var all []LBCategory
	all = append(all, LBCombat...)
	all = append(all, LBHunt...)
	all = append(all, LBMining...)
	all = append(all, LBExplore...)
	all = append(all, LBFood...)
	all = append(all, LBFails...)
	all = append(all, LBMisc...)
	all = append(all, LBExtra...)
	return all
}

type UserCacheEntry struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

func getUsernameMap() map[string]string {
	cachePath := filepath.Join(os.Getenv("MINECRAFT_PATH"), "usercache.json")
	if os.Getenv("MINECRAFT_PATH") == "" {
		cachePath = "/home/minecraft/server-minecraft/usercache.json"
	}
	
	bytes, err := ioutil.ReadFile(cachePath)
	m := make(map[string]string)
	if err != nil {
		return m
	}
	
	var cache []UserCacheEntry
	if err := json.Unmarshal(bytes, &cache); err == nil {
		for _, e := range cache {
			m[e.UUID] = e.Name
		}
	}
	return m
}

func parseStatValue(data map[string]interface{}, statID string) int {
	parts := strings.SplitN(statID, ":", 3)
	if len(parts) < 2 {
		return 0
	}
	
	statsObj, ok := data["stats"].(map[string]interface{})
	if !ok {
		return 0
	}
	
	var category, item string
	if len(parts) == 3 {
		category = parts[0] + ":" + parts[1]
		item = parts[0] + ":" + parts[2]
	} else {
		category = parts[0] + ":" + parts[1]
		item = ""
	}
	
	catObj, ok := statsObj[category].(map[string]interface{})
	if !ok {
		return 0
	}
	
	if item != "" {
		val, ok := catObj[item].(float64)
		if ok {
			return int(val)
		}
	} else {
		// Custom stats just have 2 parts usually, e.g., minecraft:custom
		// Wait, custom stats are minecraft:custom:play_time (3 parts).
		// Are there 2 part stats? Usually no. But we leave this fallback.
	}
	
	return 0
}

func GetLeaderboard(statID string) []PlayerStat {
	statsPath := filepath.Join(os.Getenv("MINECRAFT_PATH"), "world", "stats")
	if os.Getenv("MINECRAFT_PATH") == "" {
		statsPath = "/home/minecraft/server-minecraft/world/stats"
	}
	
	files, err := ioutil.ReadDir(statsPath)
	var results []PlayerStat
	if err != nil {
		return results
	}
	
	userMap := getUsernameMap()
	
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			uuid := strings.TrimSuffix(f.Name(), ".json")
			bytes, err := ioutil.ReadFile(filepath.Join(statsPath, f.Name()))
			if err != nil {
				continue
			}
			
			var data map[string]interface{}
			if err := json.Unmarshal(bytes, &data); err != nil {
				continue
			}
			
			val := parseStatValue(data, statID)
			if val > 0 {
				name := userMap[uuid]
				if name == "" {
					name = "Unknown"
				}
				results = append(results, PlayerStat{UUID: uuid, Name: name, Value: val})
			}
		}
	}
	
	sort.Slice(results, func(i, j int) bool {
		return results[i].Value > results[j].Value
	})
	
	return results
}

func GetPlayerTitles(username string) []LBCategory {
	allCats := GetAllCategories()
	var unlocked []LBCategory
	
	statsPath := filepath.Join(os.Getenv("MINECRAFT_PATH"), "world", "stats")
	if os.Getenv("MINECRAFT_PATH") == "" {
		statsPath = "/home/minecraft/server-minecraft/world/stats"
	}
	
	files, err := ioutil.ReadDir(statsPath)
	if err != nil {
		return unlocked
	}
	
	userMap := getUsernameMap()
	
	type PlayerData struct {
		Name string
		Data map[string]interface{}
	}
	var allPlayers []PlayerData
	
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			uuid := strings.TrimSuffix(f.Name(), ".json")
			bytes, err := ioutil.ReadFile(filepath.Join(statsPath, f.Name()))
			if err != nil {
				continue
			}
			var data map[string]interface{}
			if err := json.Unmarshal(bytes, &data); err == nil {
				name := userMap[uuid]
				if name != "" {
					allPlayers = append(allPlayers, PlayerData{Name: name, Data: data})
				}
			}
		}
	}
	
	for _, cat := range allCats {
		var topPlayer string
		maxVal := 0
		
		for _, p := range allPlayers {
			val := parseStatValue(p.Data, cat.StatID)
			if val > maxVal {
				maxVal = val
				topPlayer = p.Name
			}
		}
		
		if maxVal > 0 && strings.EqualFold(topPlayer, username) {
			unlocked = append(unlocked, cat)
		}
	}
	
	sort.Slice(unlocked, func(i, j int) bool {
		return unlocked[i].Stars > unlocked[j].Stars
	})
	
	if len(unlocked) > 5 {
		return unlocked[:5]
	}
	return unlocked
}

func RunRcon(cmdStr string) error {
	rconPass := os.Getenv("RCON_PASSWORD")
	rconPort := os.Getenv("RCON_PORT")
	if rconPort == "" {
		rconPort = "25575"
	}
	cmd := exec.Command("mcrcon", "-H", "127.0.0.1", "-P", rconPort, "-p", rconPass, cmdStr)
	err := cmd.Run()
	if err != nil {
		log.Printf("RCON failed for %s: %v", cmdStr, err)
	}
	return err
}
