#!/bin/bash
# Install Debian packaging tools on Arch Linux

echo "Installing Debian packaging tools on Arch Linux..."

# Install from AUR
yay -S --needed devscripts dput

echo ""
echo "✅ Installation complete!"
echo ""
echo "Next steps:"
echo "1. Set up GPG key (if you haven't):"
echo "   gpg --full-generate-key"
echo ""
echo "2. Upload key to Ubuntu keyserver:"
echo "   gpg --keyserver keyserver.ubuntu.com --send-keys YOUR_KEY_ID"
echo ""
echo "3. Add key to Launchpad:"
echo "   https://launchpad.net/~/+editpgpkeys"
echo ""
echo "4. Create PPA on Launchpad:"
echo "   https://launchpad.net/~/+activate-ppa"
echo ""
echo "5. Configure dput (create ~/.dput.cf):"
echo "   [ppa]"
echo "   fqdn = ppa.launchpad.net"
echo "   method = ftp"
echo "   incoming = ~yashsaini99/ubuntu/sdtop/"
echo "   login = anonymous"
echo ""
echo "6. Run: chmod +x packaging/debian/upload-to-ppa.sh"
echo "7. Run: ./packaging/debian/upload-to-ppa.sh"
