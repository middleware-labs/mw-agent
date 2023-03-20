Name: mw-go-agent-host-aws
Version: 0.0.1
Release: 1
Summary: My Package Summary
License: GPL
Group: Development/Tools
Source0: mw-go-agent-host-aws_0.0.1-1_all.tar.gz


%description
My package description

%prep
%setup -q

%build

%install
mkdir -p %{buildroot}/usr/bin/mw-go-agent-host-aws
cp -rfa ~/rpmbuild/BUILD/mw-go-agent-host-aws-0.0.1/* %{buildroot}/usr/bin/mw-go-agent-host-aws

%files
/usr/bin/mw-go-agent-host-aws/*
