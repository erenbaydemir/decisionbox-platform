'use client';

import { AppShell, Group, NavLink, Text, Title } from '@mantine/core';
import { IconBrain, IconFolder, IconSettings } from '@tabler/icons-react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

export default function Shell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();

  return (
    <AppShell
      navbar={{ width: 240, breakpoint: 'sm' }}
      padding="md"
    >
      <AppShell.Navbar p="md">
        <AppShell.Section>
          <Group mb="md">
            <IconBrain size={28} color="var(--mantine-color-blue-6)" />
            <Title order={4}>DecisionBox</Title>
          </Group>
        </AppShell.Section>

        <AppShell.Section grow>
          <NavLink
            component={Link}
            href="/"
            label="Projects"
            leftSection={<IconFolder size={18} />}
            active={pathname === '/'}
          />
        </AppShell.Section>

        <AppShell.Section>
          <Text size="xs" c="dimmed" ta="center">
            DecisionBox Open Source
          </Text>
        </AppShell.Section>
      </AppShell.Navbar>

      <AppShell.Main>{children}</AppShell.Main>
    </AppShell>
  );
}
