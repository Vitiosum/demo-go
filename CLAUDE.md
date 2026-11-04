# 🧠 Claude.md — demo-go

## 🏛️ Posture et méthode d'exécution

Tu es un expert cloud senior, rigoureux, structuré et orienté exécution.

Ta mission est de proposer la solution la plus cohérente, la plus pérenne et la plus simple à maintenir, avec une contrainte absolue :
- tout doit être fait exclusivement dans le cloud,
- uniquement via la console cloud,
- sans usage du local,
- sans contournement,
- sans dépendance à un poste développeur,
- sans proposer de manipulation hors plateforme.

Tu dois raisonner avec fermeté : ne propose pas plusieurs pistes floues si une option s'impose clairement. Tu analyses d'abord, tu compares rapidement les options réalistes, puis tu retiens la meilleure approche selon les critères suivants :
1. simplicité d'exploitation,
2. pérennité de l'architecture,
3. facilité d'évolution / upgrade,
4. cohérence technique,
5. faisabilité immédiate dans la console cloud,
6. réduction maximale des risques de blocage.

**Contraintes strictes :**
- ne jamais proposer de solution locale ;
- ne jamais demander d'exécuter une commande sur une machine personnelle ;
- ne jamais recommander un workflow "temporaire" si ce n'est pas industrialisable ;
- ne jamais laisser une réponse au milieu en disant "à toi de voir" ou "choisis parmi ces options" ;
- tu dois trancher et recommander une solution principale ;
- si une idée n'est pas compatible avec une exécution 100 % cloud console, tu l'écartes explicitement ;
- tu privilégies la solution la plus robuste et la plus simple à reprendre plus tard.

**Méthode de réponse obligatoire :**
1. Reformuler brièvement le besoin.
2. Identifier les contraintes bloquantes.
3. Lister les options réellement possibles dans le cadre 100 % cloud console.
4. Écarter clairement les mauvaises options avec justification.
5. Retenir une seule stratégie recommandée.
6. Donner un plan d'exécution concret, ordonné, sans trous.
7. Préciser les points de vigilance.
8. Donner le résultat attendu une fois la mise en place terminée.

**Format attendu :** Réponse structurée, phrases claires, ton ferme, professionnel, décisionnel. Pas de blabla, pas d'hésitation, pas de théorie inutile, pas de proposition hors périmètre.

> Toute recommandation doit être pensée pour être durable, propre techniquement, et directement applicable dans le cloud sans blocage ni dépendance cachée.

---

## 🎯 Contexte du projet

Aquarium du runtime Go en temps réel (refonte visuelle du 2026-11-04, après l'oscilloscope du tournage vidéo).
Un poisson par goroutine (dorés : goroutines de fond ; banc corail : la rafale en cours), l'eau au niveau du tas avec la ligne rouge `HEAP_ALERT_MB` gravée sur la vitre, la pompe du GC sur le côté (rotor, passage du filtre, compteur de cycles), une bulle par pause de GC étiquetée en µs. Plaque en laiton pour les vrais chiffres, enregistreur des 90 dernières secondes, pupitre du soigneur à trois boutons (charge bornée).
Certification Clever Cloud Academy dans un hublot en laiton (badge, « Cette app tourne sur Clever Cloud. Vous aussi, certifiez-vous. », deux parcours), plus le bouton « Se certifier » de la topbar et un lien depuis l'intro.

Runtime cible : **Go** sur Clever Cloud. Aucune base de données.

---

## 🎨 Design system — Clever Brand Kit + univers « aquarium »

**Kit :** `static/cc-brand.css`, copie **à l'identique** du kit partagé. Ne pas le modifier : les styles propres à la démo sont dans `static/aquarium.css` (préfixe `.aq-`, plus `.go-announce` et `.go-env`).
**Typographie :** Marcellus (plaques gravées, titres), Plus Jakarta Sans (texte), JetBrains Mono (compteurs de la plaque, noms de variables, règle du bassin). Google Fonts seulement.
**Couleurs :** hall navy `#13172e` carrelé ; eau `#8fe3f3 → #2bb3d6 → #0a4a6a` ; banc corail `#f57461` ; résidents dorés `#f2c14e` ; laiton `#c9a35b` ; alerte `#ff5a72` ; GC `#5fe3d3`. Texte blanc sur dégradé = départ assombri `#cc4432 → #cb1c42 → #a51050` (AA).
**Structure de la page :**
1. `.cc-topbar` — logo, « Go Runtime », pill « Go x.y · stdlib », pill live/local, « Se certifier ↗ »
2. `.go-announce` — « Avis aux visiteurs » + `ANNOUNCEMENT` (si non vide)
3. `.aq-scene` — titre « N goroutines nagent en ce moment. », intro (lien `#cert-title`) ; `.aq-hood` (nom d'app + pastille `APP_ENV`) ; `.aq-tank` (`<canvas id="tank">` dessiné par `aquarium.js`, pastille `#over`) ; `.aq-stand` = plaque en laiton (goroutines, tas, cycles, pause, GOMAXPROCS, uptime, pages), pupitre du soigneur (3 boutons), enregistreur 90 s (SVG)
4. `.aq-tags` — « Réglages du bassin » (variables de démo) et « Running on Clever Cloud » (variables plateforme)
5. `.aq-cert` — hublot `.aq-porthole` avec le badge (`{{template "cc-badge"}}`, une seule fois par page : il porte des `id`), deux parcours Academy, « Passer la certification »
6. `.aq-build` — « Comment le bassin a été rempli » (4 étapes réelles du build Clever, en canalisation)
7. `.cc-footer`

**Échelle du bassin :** `niceUp(max(HEAP_ALERT_MB × 1.25, 64 × 1.15, pic du tas × 1.1))` → 80 MB par défaut, 200 MB avec `HEAP_ALERT_MB=128`. Le 0 MB de la règle est au tiers bas : le bassin n'est jamais à sec.
**Motion :** poissons, vagues, plantes, pompe et bulles animés en `requestAnimationFrame` ; badge en rotation lente ; `.cc-reveal` du kit. Sous `prefers-reduced-motion`, le bassin est redessiné immobile à chaque mesure.

---

## ☁️ Déploiement Clever Cloud

- **Type d'app** : `go` (`clever create --type go`), aucun add-on, pas de build dédié nécessaire
- **Build** : méthode `gomod` — `go install` du module `github.com/Vitiosum/demo-go` lu dans `go.mod`, binaire `$GOPATH/bin/demo-go` lancé tel quel
- **Version Go** : `go.mod` déclare `go 1.27` ; l'image Clever est en Go 1.27.3 (changelog « Images update » du 2026-10-21). Ne jamais déclarer une version plus récente que l'image : la doc indique « If the Go version declared in your go.mod file is newer than the one installed on the instance, the Go toolchain downloads the required version automatically » (téléchargement à chaque build). `CC_GO_VERSION` n'existe pas.
- **`clevercloud/go.json`** : déprécié (« Using clevercloud/go.json to define Makefile and binary paths is a deprecated method and should no longer be used »). Son effet réel au build n'a pas été mesuré ; sa suppression reste une décision utilisateur.
- **Port** : `PORT` ou `8080` ; écoute sur `0.0.0.0`
- **Routes** : `GET /`, `GET /stats` (`?history=1` pour les 90 dernières mesures ; chaque mesure `{t, g, b, h, gc, p}`, `b` = goroutines de la rafale encore vivantes), `GET /health`, `GET /cc-brand.css`, `GET /aquarium.css`, `GET /aquarium.js`, `GET /app.js`, `POST /load/goroutines|heap|gc` — autres verbes → 405, autres chemins → 404

### Variables Clever à poser (vérifiées dans la doc le 2026-10-25)
| Variable | Valeur | Preuve doc |
|---|---|---|
| `CC_GO_BUILD_TOOL` | `gomod` | « Available values: gomod, gobuild, makefile. Determines how to build and install your application. » ; gomod « Requires a go.mod file at the root of your application » |
| `CC_HEALTH_CHECK_PATH` | `/health` | « If these paths don't return a 2xx response code, the deployment fails. » |

Inutiles ici : `CC_GO_PKG` (« Overrides auto-detection from go.mod », le module est à la racine), `CC_GO_RUNDIR` (« Deprecated »), `CC_CACHE_DEPENDENCIES` (« Enable caching of your build dependencies » : aucune dépendance à mettre en cache), `CC_IGNORE_FROM_BUILDCACHE` (rien de lourd dans l'archive). Le build compile la seule bibliothèque standard : c'est déjà le cas le plus court.

### Variables de démo (onglet Variables de la console)
| Variable | Défaut | Effet visible |
|---|---|---|
| `APP_ENV` | `local` | Pastille sur le capot du bassin, dégradé Clever si `production` |
| `ANNOUNCEMENT` | vide | Bandeau « Avis aux visiteurs » en haut de la page (160 caractères max) |
| `BURST_GOROUTINES` | `500` (1..2000) | Libellé du bouton, taille du banc de poissons corail, pic sur l'enregistreur |
| `HEAP_ALERT_MB` | `32` (1..1024) | Ligne rouge gravée sur la vitre et sur l'enregistreur ; au-dessus, chiffre du tas en rouge et « Heap au-dessus de N MB » |

Valeur invalide ou hors bornes → défaut (`loadDemoConfig`, testé).
Variables plateforme affichées (« injecté au déploiement » en local) : `CC_APP_NAME`, `APP_ID` (bascule Live/Local), `INSTANCE_NUMBER` · `CC_PRETTY_INSTANCE_NAME`, `INSTANCE_TYPE`, `COMMIT_ID`/`CC_COMMIT_ID` (7 car.), `CC_DEPLOYMENT_ID`/`DEPLOYMENT_ID` (16 car.).

> **Commit déployé** : lu dans `COMMIT_ID`/`CC_COMMIT_ID`. Vérifié sur demo-go le 2026-11-04 : `CC_COMMIT_ID` est posé par `clever deploy` et reste renseigné après un `clever restart`. Si la variable est vide sur la plateforme, l'étiquette « Running on Clever Cloud » affiche « non renseigné ».

### Charge bornée (pupitre)
- Goroutines : `BURST_GOROUTINES` goroutines qui dorment 4 à 12 s ; une rafale à la fois (CAS atomique), sinon 429 + `Retry-After`
- Heap : 64 MB fixes (pages touchées), tenus 8 s, libérés puis `runtime.GC()` ; une allocation à la fois
- GC : `runtime.GC()` au plus toutes les 2 s
- Pire cas : 2000 goroutines endormies + 64 MB. Aucune boucle CPU.

### Scénario vidéo
1. Ouvrir la page : quelques poissons dorés (goroutines de fond), l'eau au repos ; l'enregistreur est déjà plein (historique servi par le serveur).
2. « Lâcher un banc de N » : un banc corail jaillit sous la surface ; le titre passe à plusieurs centaines, puis le banc s'éclaircit en 4 à 12 s.
3. « Verser 64 MB dans le tas » : l'eau monte au-dessus de la ligne rouge, chiffre rouge et pastille d'alerte ; 8 s plus tard, la pompe tourne, l'eau redescend, une bulle de pause remonte.
4. Console Clever → Variables : `HEAP_ALERT_MB=128`, `ANNOUNCEMENT=…`, `APP_ENV=production` ; `clever restart` ; recharger : avis aux visiteurs, pastille dégradée, règle jusqu'à 200 MB (le versement ne franchit plus la ligne).
5. Défiler vers le hublot de la certification → « Passer la certification ».

---

## 🛠️ Stack

| Élément | Valeur |
|---|---|
| Go | 1.27 (`go.mod`, sans directive toolchain) |
| Dépendances | Aucune (stdlib + `embed`) |
| Frontend | `static/index.html` (`html/template`), `static/aquarium.js` (canvas), `static/app.js`, `static/aquarium.css`, `static/cc-brand.css`, embarqués dans le binaire |
| Base de données | Aucune |

---

## 📁 Structure clé

```
main.go               → serveur HTTP (timeouts, arrêt propre SIGTERM), routes, middleware secure() (CSP), handlers
pulse.go              → échantillonneur 1 mesure/s (anneau de 90, compteur des goroutines de rafale), générateurs de charge bornés
config.go             → variables de démo (loadDemoConfig) et variables Clever (loadPlatform)
main_test.go          → tests httptest (routes, en-têtes, variables de démo, bornes de charge)
static/index.html     → template de la page (logo et badge SVG en {{define}})
static/aquarium.js    → le bassin sur <canvas> : poissons, eau, pompe, bulles (fichier séparé : CSP)
static/app.js         → polling /stats, plaque en laiton, pupitre, enregistreur 90 s
static/aquarium.css   → styles de l'univers aquarium
static/cc-brand.css   → kit CSS partagé (copie, ne pas modifier)
clevercloud/go.json   → ancien fichier, déprécié par la doc Clever (suppression : décision utilisateur)
```

---

## ⚙️ Commandes utiles

```bash
# Lancer en local (port de tournage : 8104)
PORT=8104 go run .
PORT=8104 APP_ENV=production ANNOUNCEMENT="Maintenance ce soir à 22 h" BURST_GOROUTINES=800 HEAP_ALERT_MB=40 go run .

# Vérifier / tester
gofmt -l . ; go vet ./... && go test -race ./...

# Vulnérabilités connues (stdlib incluse)
govulncheck ./...    # go install golang.org/x/vuln/cmd/govulncheck@latest
```

---

## 🚀 Déployer une modification

```bash
git add .
git commit -m "description"
git push          # sauvegarde GitHub uniquement
clever deploy     # déploie main sur Clever Cloud
```

Pas de liaison GitHub → Clever Cloud (`clever create --github` refusé par l'API, « wrong oauth application data » ; intégration GitHub du compte à reconnecter dans la Console) : un `git push` ne redéploie pas.

App en service du 2026-10-25 au 2026-11-05 (suppression planifiée après le tournage vidéo du 2026-11-04) : `app_61a7ff90-36fe-49eb-b42b-2077d0e9db58`, https://app-61a7ff90-36fe-49eb-b42b-2077d0e9db58.cleverapps.io/. Après cette date, recréer avec la procédure du README.

---

## ⚠️ Points de vigilance

- Le HTML vit dans `static/index.html`, embarqué par `//go:embed static` : toujours builder le **package** (`go run .`, `go build .`), pas un fichier isolé
- L'app écoute sur `0.0.0.0:PORT` — ne pas hardcoder le port
- `INSTANCE_NUMBER` est injecté par Clever Cloud pour distinguer les instances en cas de scaling
- `static/cc-brand.css` est une copie du kit partagé : le remplacer entièrement lors d'une mise à jour du kit, ne pas l'éditer localement
- Le template est un `html/template` : les valeurs sont échappées automatiquement, ne pas y injecter de HTML brut
- CSP stricte (`script-src 'self'`) : tout JavaScript va dans `static/aquarium.js` ou `static/app.js`, jamais en `<script>` inline ni en attribut `on*` ; seules Google Fonts sont autorisées en externe
- `runtime.ReadMemStats` arrête le monde : il ne tourne que dans l'échantillonneur, jamais dans un handler
- Toute nouvelle charge du pupitre doit rester bornée (une à la fois, taille fixe, 429 sinon)
- Le serveur a des timeouts (lecture 10 s, écriture 15 s) et s'arrête proprement sur SIGTERM (10 s) : ne pas revenir à `http.ListenAndServe` nu

---

## 🔍 Diagnostic rapide

| Symptôme | Cause probable | Correction |
|---|---|---|
| App ne démarre pas | Port non écouté sur 0.0.0.0 | Vérifier que `ListenAndServe` utilise `0.0.0.0:PORT` |
| `pattern static: no matching files` | Build hors du package | Utiliser `go run .` / `go build .` |
| Page sans style | `/cc-brand.css` en 404 | Vérifier que `static/cc-brand.css` est présent et committé |
| Lien doc cassé | URL Clever Cloud modifiée | Mettre à jour l'URL dans `static/index.html` |
| Métriques figées | Erreur fetch `/stats` côté JS | Vérifier les logs runtime Clever Cloud ; le pupitre affiche « Hors ligne, nouvelle tentative… » |
| Bouton sans effet, « Une rafale tourne déjà » | Rafale précédente encore active (429) | Attendre le compte à rebours affiché sur le bouton |
| Build plus long que prévu | `go.mod` déclare une version plus récente que l'image | Revenir à une version ≤ celle de l'image (changelog Clever) |
| Script bloqué (console : « Refused to execute ») | Script inline ou domaine hors CSP | Mettre le JS dans `static/` (servi par une route de `main.go`) ou étendre `contentSecurityPolicy` |
| « injecté au déploiement » partout | `APP_ID` absent | Normal en local ; renseigné automatiquement sur Clever Cloud |
