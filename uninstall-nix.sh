#!/bin/bash
# uninstall-nix.sh — Remove nix and nix-darwin from macOS
# Run with: sudo bash uninstall-nix.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}This script must be run with sudo:${NC}"
  echo "  sudo bash $0"
  exit 1
fi

echo -e "${YELLOW}╔══════════════════════════════════════════╗${NC}"
echo -e "${YELLOW}║   Nix / nix-darwin Uninstaller for macOS ║${NC}"
echo -e "${YELLOW}╚══════════════════════════════════════════╝${NC}"
echo ""
echo "This will remove:"
echo "  • Nix daemon and LaunchDaemons"
echo "  • Nix build users and group"
echo "  • /nix APFS volume"
echo "  • /etc/nix config"
echo "  • Shell hooks"
echo ""
read -p "Continue? [y/N] " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
  echo "Cancelled."
  exit 0
fi

echo ""
echo -e "${GREEN}[1/7] Stopping nix LaunchDaemons...${NC}"
for svc in nix-daemon darwin-store activate-system nix-gc nix-optimise; do
  if launchctl print system/org.nixos.$svc &>/dev/null; then
    launchctl bootout system/org.nixos.$svc 2>/dev/null || true
    echo "  ✓ stopped org.nixos.$svc"
  fi
done
rm -f /Library/LaunchDaemons/org.nixos.*.plist
echo "  ✓ removed plist files"

echo ""
echo -e "${GREEN}[2/7] Removing nix build users...${NC}"
for u in $(dscl . -list /Users | grep _nixbld); do
  dscl . -delete /Users/$u
  echo "  ✓ removed $u"
done
if dscl . -read /Groups/nixbld &>/dev/null; then
  dscl . -delete /Groups/nixbld
  echo "  ✓ removed nixbld group"
fi

echo ""
echo -e "${GREEN}[3/7] Removing /nix APFS volume...${NC}"
# Kill any processes running from /nix (exclude this script)
ps aux | grep '/nix/store' | grep -v grep | grep -v "uninstall-nix" | awk '{print $2}' | xargs kill -9 2>/dev/null || true
sleep 1
if mount | grep -q "on /nix"; then
  if ! diskutil apfs deleteVolume /nix 2>/dev/null; then
    echo "  – Normal delete failed, trying force..."
    if ! diskutil apfs deleteVolume /nix -force 2>/dev/null; then
      echo -e "  ${YELLOW}⚠  Volume held by kernel. Will be deleted after reboot.${NC}"
      NEEDS_POST_REBOOT_DELETE=1
    else
      echo "  ✓ force-deleted /nix volume"
    fi
  else
    echo "  ✓ deleted /nix volume"
  fi
else
  echo "  – /nix not mounted, skipping"
fi

echo ""
echo -e "${GREEN}[4/7] Cleaning /etc/synthetic.conf...${NC}"
if [ -f /etc/synthetic.conf ]; then
  sed -i '' '/^nix$/d' /etc/synthetic.conf
  echo "  ✓ removed nix entry"
  if [ ! -s /etc/synthetic.conf ]; then
    rm /etc/synthetic.conf
    echo "  ✓ removed empty synthetic.conf"
  fi
fi

echo ""
echo -e "${GREEN}[5/7] Removing /etc/nix...${NC}"
rm -rf /etc/nix
echo "  ✓ done"

echo ""
echo -e "${GREEN}[6/7] Removing nix shell hooks...${NC}"
for f in /etc/zshrc /etc/bashrc /etc/bash.bashrc; do
  if [ -f "$f" ] && grep -q nix "$f"; then
    sed -i '' '/nix/d' "$f"
    echo "  ✓ cleaned $f"
  fi
done
rm -f /etc/profile.d/nix.sh 2>/dev/null
# Clean user shell config
REAL_USER="${SUDO_USER:-$USER}"
REAL_HOME=$(eval echo "~$REAL_USER")
if grep -q nix "$REAL_HOME/.zshrc" 2>/dev/null; then
  sed -i '' '/nix/d' "$REAL_HOME/.zshrc"
  echo "  ✓ cleaned $REAL_HOME/.zshrc"
fi

echo ""
echo -e "${GREEN}[7/7] Cleanup...${NC}"
rm -rf /nix 2>/dev/null || true
rm -f /etc/nix-darwin 2>/dev/null || true
echo "  ✓ done"

echo ""
echo -e "${GREEN}════════════════════════════════════════════${NC}"
echo -e "${GREEN}  Nix has been removed.${NC}"
echo -e "${GREEN}════════════════════════════════════════════${NC}"
echo ""
if [ "${NEEDS_POST_REBOOT_DELETE:-0}" = "1" ]; then
  echo -e "${YELLOW}⚠  After reboot, run this to delete the volume:${NC}"
  echo "     sudo diskutil apfs deleteVolume disk1s7"
  echo ""
fi
echo -e "${YELLOW}⚠  You must reboot to clear the /nix firmlink.${NC}"
echo ""
read -p "Reboot now? [y/N] " reboot_confirm
if [[ "$reboot_confirm" == "y" || "$reboot_confirm" == "Y" ]]; then
  reboot
fi
