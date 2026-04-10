# Design Spec — Aura Full Redesign · demo-go

**Date:** 2026-04-10  
**Projet:** demo-go (Go Runtime Dashboard)  
**Statut:** Approuvé

---

## Contexte

Le dashboard Go runtime passe du design system "Track Night" (Bebas Neue, orange #FF5A1F, fond #07080C) au design system **Aura AI Automation** (Inter + Newsreader, bleu #3b82f6, fond hsl(0,0%,9%)).

L'app est un seul fichier `main.go` — tout le HTML/CSS/JS est embarqué dans la constante `indexHTML`. Le redesign s'applique exclusivement à cette constante. Aucune dépendance externe n'est ajoutée côté Go.

---

## Approche retenue : Aura Full

Tous les composants du design system Aura sont implémentés :

- Nav glass blur (sticky header avec backdrop-filter)
- Orbs iridescents (bleu/violet/cyan, positionnés en fixed)
- Hero Inter 700 + Newsreader italic 300
- Cartes métriques avec top accent 2px couleur
- Status bar verte
- Marquee avec gradient mask
- Shiny CTA (conic-gradient border animé) pour la certification
- Footer liens colorés

---

## Couleurs

| Rôle | Valeur |
|---|---|
| Background principal | `hsl(0, 0%, 9%)` — #171717 |
| Background cartes/surfaces | `hsl(0, 0%, 11%)` — #1c1c1c |
| Texte principal | `hsl(0, 0%, 98%)` — #fafafa |
| Texte muted | `hsl(0, 0%, 65%)` — #a6a6a6 |
| Border | `hsl(0, 0%, 20%)` — #333333 |
| Accent bleu (primary) | `#3b82f6` / `#2563eb` |
| Accent or (CTA certif) | `#fbbf24` / `#f59e0b` |
| Top accent violet (heap) | `#8b5cf6` |
| Top accent vert (uptime) | `#22c55e` |
| Top accent cyan (GC) | `#06b6d4` |
| Top accent slate (version) | `hsl(0,0%,40%)` |

---

## Typographie

| Usage | Famille | Poids | Letter-spacing |
|---|---|---|---|
| H1 hero | Inter | 700 | -0.05em |
| Sous-titre hero | Newsreader | italic 300 | -0.02em |
| Labels cartes | Inter | 600 | +0.1em (uppercase) |
| Valeurs métriques | DM Mono | 300 | -0.02em |
| Corps / navigation | Inter | 400–600 | -0.01em |

**Chargement polices :** Google Fonts — Inter (300,400,600,700,italic), Newsreader (italic 300), DM Mono (300,400).

---

## Composants

### 1. Orbs iridescents (background)
- 3 orbes en `position:fixed`, `pointer-events:none`, `filter:blur(60px)`, `opacity:0.25`
- Orb 1 : bleu `#3b82f6`, haut droite (400×400px)
- Orb 2 : violet `#8b5cf6`, bas gauche (300×300px)
- Orb 3 : cyan `#06b6d4`, centre (200×200px)
- Pas d'animation CSS (évite la surcharge dans `main.go`) — positionnement statique suffit pour l'effet

### 2. Nav glass blur (sticky)
- `position:sticky; top:0; z-index:10`
- `backdrop-filter:blur(16px)` + `background:rgba(23,23,23,0.7)`
- `border-bottom:1px solid hsl(0,0%,20%)`
- Logo "Go.Runtime" (Inter 700, le point en `#3b82f6`)
- Badge "LIVE" à droite : pill avec dot bleu pulsant

### 3. Hero
- Badge "Live runtime" : pill bleu pulsant (même dot que nav)
- Titre : `font-size:clamp(2.5rem,7vw,4rem)`, Inter 700, letter-spacing `-0.05em`
- "dashboard" en dessous en Newsreader italic 300, `color:hsl(0,0%,65%)`
- Sous-titre : Inter 400, muted

### 4. Cartes métriques (6 cartes, grille 3×2)
- Background `hsl(0,0%,11%)`, border `1px solid hsl(0,0%,20%)`, border-radius `12px`
- Barre top 2px colorée (gradient) par carte :
  - Goroutines → bleu
  - Heap → violet
  - Uptime → vert
  - GC Cycles → cyan
  - Requests → or
  - Go Version → slate
- Label : 9px, uppercase, letter-spacing +0.1em, muted
- Valeur : DM Mono 300, 28px (sauf Version : 20px)
- Unité : 10px, muted
- Flash animation au changement (identique à l'actuel — radial-gradient + opacity)

### 5. Status bar
- Dot vert `#22c55e` pulsant + texte "LIVE · updates every 2s · Last: HH:MM:SS"
- Inter 11px, muted, centré

### 6. Info row (Host / Instance / Port)
- Conteneur : `hsl(0,0%,11%)`, border, border-radius 12px, padding
- 3 sous-cases : background `hsl(0,0%,9%)`, border, border-radius 8px
- Label 8px uppercase, valeur DM Mono 11px

### 7. Marquee
- Conteneur avec `overflow:hidden` + `position:relative`
- Masques gradient gauche/droite : `linear-gradient(90deg, hsl(0,0%,9%), transparent)`
- Items : background `hsl(0,0%,11%)`, border, border-radius 6px, Inter 10px, muted
- Contenu : "Go 1.24 · Clever Cloud · Goroutines · Runtime Metrics · GC · stdlib only · No dependencies · Open source" × 2 (boucle)
- Animation : `@keyframes marquee` avec `transform:translateX(-50%)` en 30s linéaire infini

### 8. Shiny CTA (certification)
- Pill-shape, `border-radius:99px`
- Border via `::before` avec `conic-gradient(from 0deg, #3b82f6, #8b5cf6, #06b6d4, #fbbf24, #3b82f6)`
- Animation `@keyframes border-spin` : rotate 360deg en 4s linéaire infini sur `::before`
- Background intérieur : `hsl(0,0%,9%)` avec dot-pattern `radial-gradient(circle, rgba(255,255,255,0.07) 1px, transparent 1px)` en `background-size:16px 16px`
- Icône 🎓 + texte "Obtenir la certification Clever Cloud →"
- Lien vers `https://academy.clever.cloud/`

### 9. Footer
- Border-top `hsl(0,0%,18%)`
- Liens pill identiques à l'actuel (clever-cloud.com vert, LinkedIn bleu, Go on CC, Go Docs, Certification or)
- Copyright inchangé

---

## Structure HTML (ordre dans `indexHTML`)

```
<head> Fonts Google + <style> CSS complet
<body>
  <!-- Orbs -->
  .orb × 3

  <!-- Nav -->
  <nav class="nav">

  <!-- Container -->
  <div class="container">
    <!-- Hero -->
    .hero > .live-badge + h1 + .hero-sub-serif + .hero-sub

    <!-- Cards -->
    .metrics > .card × 6

    <!-- Status -->
    .status-bar

    <!-- Info -->
    .info-card > .info-grid

    <!-- Marquee -->
    .marquee-wrap > .marquee-track

    <!-- Shiny CTA -->
    .shiny-wrap > a.shiny-cta

    <!-- Footer -->
    <footer>
  </div>

  <!-- Script JS fetch /stats -->
</body>
```

---

## JavaScript (inchangé fonctionnellement)

- `fetchStats()` : poll `/stats` toutes les 2s
- `tickUptime()` : interpolation uptime côté client (1s)
- `flash()` : animation flash carte au changement de valeur
- Aucune dépendance JS externe

---

## Contraintes Clever Cloud

- Tout reste dans `main.go` — pas de fichiers statiques séparés
- Polices chargées depuis Google Fonts CDN (pas de téléchargement local)
- Port via `os.Getenv("PORT")` — inchangé
- `INSTANCE_NUMBER` injecté par Clever Cloud — inchangé
- Build : `clevercloud/go.json` avec `appIsToBeBuilt:true` — inchangé

---

## Résultat attendu

Le dashboard Go runtime affiche :
- Un fond sombre Aura avec 3 orbes lumineux bleu/violet/cyan
- Une nav sticky glass blur avec logo et badge LIVE
- Un hero Inter/Newsreader élégant
- 6 cartes métriques avec top accent coloré et valeurs DM Mono
- Un marquee défilant les tags Go / Clever Cloud
- Un bouton Shiny CTA animé pour la certification
- Un footer sobre avec liens colorés

Déploiement : `git push` → Clever Cloud rebuild automatique Go.
