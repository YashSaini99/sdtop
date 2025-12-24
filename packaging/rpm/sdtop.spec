Name:           sdtop
Version:        1.0.0
Release:        1%{?dist}
Summary:        Terminal-based systemd service manager

License:        MIT
URL:            https://github.com/YashSaini99/sdtop
Source0:        https://github.com/YashSaini99/%{name}/archive/v%{version}.tar.gz

BuildRequires:  golang >= 1.21
BuildRequires:  systemd-devel
Requires:       systemd

%description
sdtop is an interactive terminal application for managing systemd services
with real-time log streaming, process tree visualization, and intuitive
keyboard-driven interface.

%prep
%autosetup

%build
export CGO_ENABLED=1
export CGO_CFLAGS="${CFLAGS}"
export CGO_LDFLAGS="${LDFLAGS}"
go build -ldflags="-s -w" -o %{name} ./cmd/main.go

%install
install -Dm755 %{name} %{buildroot}%{_bindir}/%{name}
install -Dm644 README.md %{buildroot}%{_docdir}/%{name}/README.md

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}

%changelog
* Tue Dec 24 2025 Your Name <your.email@example.com> - 1.0.0-1
- Initial release
