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

Dashboard de métriques runtime Go en temps réel.
L'app expose ses propres métriques Go (goroutines, heap, GC, uptime, requests) via une interface HTML mise à jour toutes les 2 secondes par polling JavaScript, et affiche ce que la plateforme injecte (panneau « Vu depuis Clever Cloud »).
Conçue comme démo de déploiement sur **Clever Cloud**, avec la **certification Clever Cloud** (Academy) en position centrale.

Déployée sur **Clever Cloud** (runtime Go).

---

## 🎨 Design system — Clever Brand Kit

**Kit :** `static/cc-brand.css`, copie **à l'identique** du kit partagé (`docs/superpowers/brand-kit/cc-brand.css` du monorepo). Ne pas le modifier : les styles propres à la démo tiennent dans le `<style>` court de `static/index.html` (`.go-layout`, `.go-metrics`, `.go-live`).
**Spec :** `docs/superpowers/specs/2026-09-06-clever-brand-design.md`
**Typographie :** Plus Jakarta Sans (texte), JetBrains Mono (valeurs techniques) — Google Fonts, repli système.
**Couleurs :** fond navy `#13172e`, surfaces `#1c2045`, texte `#f9f9fb`, dégradé Clever `#f57461 → #cb1c42 → #a51050`, vert `#11bea9` (ok), orange `#f57461` (live, flash).
**Structure de page (ordre imposé par `reference.html`) :**
1. `.cc-topbar` — logo Clever Cloud (SVG inline, `{{template "cc-logo"}}`), nom « Go Runtime », pill mono « Go {{.GoVer}} · stdlib », pill live/local, « Se certifier ↗ »
2. `.cc-hero` — « Go, *sans dépendance, en production.* », lead, 4 facts
3. `.cc-cert` — badge SVG inline, 2 parcours Academy, CTA « Obtenir ma certification »
4. Contenu démo — 6 `.cc-card--accent` avec `.cc-stat` (`id` `v-*` lus par le JS) + `.cc-platform` « Vu depuis Clever Cloud »
5. « Ce que Clever Cloud fait pour cette app » — 3 cartes
6. `.cc-footer` — doc Go, variables d'environnement, Console, Academy (`is-cert`), GitHub

**Motion :** point live pulsant, `.cc-reveal` (fade-in échelonné), flash `.cc-stat--flash` 300 ms sur changement de valeur. Pas d'orbes, pas de marquee, pas de bordure conique.

**Dernier redesign :** Clever Brand Kit — branche `redesign/clever-brand`

---

## ☁️ Déploiement Clever Cloud

- **Type d'app** : Go
- **Config** : `clevercloud/go.json` → `appIsToBeBuilt: true`
- **Port** : `PORT` (env var Clever Cloud) ou `8080` par défaut
- **Endpoints** : `GET /` (dashboard), `GET /health`, `GET /stats` (JSON), `GET /cc-brand.css`

### Variables d'environnement
Aucune variable spécifique requise. Clever Cloud injecte `PORT` et `INSTANCE_NUMBER` automatiquement.
Le panneau plateforme lit aussi `CC_APP_NAME`, `APP_ID`, `INSTANCE_TYPE`, `CC_PRETTY_INSTANCE_NAME`, `CC_COMMIT_ID` (7 car.), `CC_DEPLOYMENT_ID` (16 car.) ; sans `APP_ID` il affiche « Local · hors Clever Cloud ».

---

## 🛠️ Stack

| Élément | Valeur |
|---|---|
| Go | 1.24 (`go.mod`) |
| Dépendances | Aucune (stdlib + `embed`) |
| Frontend | `static/index.html` (`html/template`) + `static/cc-brand.css`, embarqués dans le binaire |
| Base de données | Aucune |

---

## 📁 Structure clé

```
main.go               → serveur HTTP, /stats, /health, /cc-brand.css, platformInfo() (variables CC_*)
static/index.html     → template de la page (Clever Brand Kit) + JS de polling
static/cc-brand.css   → kit CSS partagé (copie, ne pas modifier)
clevercloud/go.json   → config de build Clever Cloud
docs/superpowers/     → specs et plans
go.mod                → module Go
```

---

## ⚙️ Commandes utiles

```bash
# Lancer en local
go run .
PORT=8084 go run .

# Vérifier / builder
go vet ./... && go build -o app .
```

---

## 🚀 Déployer une modification

```bash
git add .
git commit -m "description"
git push
```

Clever Cloud redéploie automatiquement après chaque push.

---

## ⚠️ Points de vigilance

- Le HTML vit dans `static/index.html`, embarqué par `//go:embed static` : toujours builder le **package** (`go run .`, `go build .`), pas un fichier isolé
- L'app écoute sur `0.0.0.0:PORT` — ne pas hardcoder le port
- `INSTANCE_NUMBER` est injecté par Clever Cloud pour distinguer les instances en cas de scaling
- `static/cc-brand.css` est une copie du kit partagé : le remplacer entièrement lors d'une mise à jour du kit, ne pas l'éditer localement
- Le template est un `html/template` : les valeurs sont échappées automatiquement, ne pas y injecter de HTML brut

---

## 🔍 Diagnostic rapide

| Symptôme | Cause probable | Correction |
|---|---|---|
| App ne démarre pas | Port non écouté sur 0.0.0.0 | Vérifier que `ListenAndServe` utilise `0.0.0.0:PORT` |
| `pattern static: no matching files` | Build hors du package | Utiliser `go run .` / `go build .` |
| Page sans style | `/cc-brand.css` en 404 | Vérifier que `static/cc-brand.css` est présent et committé |
| Lien doc cassé | URL Clever Cloud modifiée | Mettre à jour l'URL dans `static/index.html` |
| Métriques figées | Erreur fetch `/stats` côté JS | Vérifier les logs runtime Clever Cloud ; la pill passe en « Hors ligne » |
| Panneau plateforme vide | `APP_ID` absent | Normal en local ; renseigné automatiquement sur Clever Cloud |
