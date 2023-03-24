Name: mw-go-agent-host-aws-arm
Version: 0.0.1
Release: 1
Summary: My Package Summary
License: GPL
Group: Development/Tools
Source0: mw-go-agent-host-aws-arm_0.0.1-1_arm64.tar.gz


%description
My package description

%prep
%setup -q

%build

%install
mkdir -p %{buildroot}/usr/bin/mw-go-agent-host-aws-arm
cp -rfa ~/rpmbuild/BUILD/mw-go-agent-host-aws-arm-0.0.1/* %{buildroot}/usr/bin/mw-go-agent-host-aws-arm

%files
/usr/bin/mw-go-agent-host-aws-arm/*
