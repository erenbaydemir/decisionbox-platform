'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import {
  Badge, Button, Card, Grid, Group, Loader, Stack, Tabs, Text, Title,
} from '@mantine/core';
import { notifications } from '@mantine/notifications';
import {
  IconAlertTriangle, IconBulb, IconChartBar, IconPlayerPlay, IconTrendingUp,
} from '@tabler/icons-react';
import Shell from '@/components/layout/AppShell';
import { api, DiscoveryResult, Insight, Project, Recommendation } from '@/lib/api';

const severityColor: Record<string, string> = {
  critical: 'red', high: 'orange', medium: 'yellow', low: 'gray',
};

export default function ProjectPage() {
  const { id } = useParams<{ id: string }>();
  const [project, setProject] = useState<Project | null>(null);
  const [discovery, setDiscovery] = useState<DiscoveryResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [triggering, setTriggering] = useState(false);

  useEffect(() => {
    Promise.all([
      api.getProject(id).then(setProject),
      api.getLatestDiscovery(id).then(setDiscovery).catch(() => null),
    ])
      .catch((e) => notifications.show({ title: 'Error', message: e.message, color: 'red' }))
      .finally(() => setLoading(false));
  }, [id]);

  const handleTrigger = async () => {
    setTriggering(true);
    try {
      const result = await api.triggerDiscovery(id);
      notifications.show({ title: 'Discovery triggered', message: result.message, color: 'blue' });
    } catch (e: unknown) {
      notifications.show({ title: 'Error', message: (e as Error).message, color: 'red' });
    } finally {
      setTriggering(false);
    }
  };

  if (loading) {
    return <Shell><Loader /></Shell>;
  }

  if (!project) {
    return <Shell><Text>Project not found</Text></Shell>;
  }

  // Group insights by analysis area
  const insightsByArea: Record<string, Insight[]> = {};
  discovery?.insights?.forEach((insight) => {
    if (!insightsByArea[insight.analysis_area]) {
      insightsByArea[insight.analysis_area] = [];
    }
    insightsByArea[insight.analysis_area].push(insight);
  });

  const areas = Object.keys(insightsByArea);

  return (
    <Shell>
      <Stack gap="lg">
        {/* Header */}
        <Group justify="space-between">
          <div>
            <Title order={2}>{project.name}</Title>
            <Group gap="xs" mt={4}>
              <Badge variant="light">{project.domain}</Badge>
              <Badge variant="light" color="blue">{project.category}</Badge>
              <Badge variant="light" color={project.status === 'active' ? 'green' : 'gray'}>
                {project.status}
              </Badge>
            </Group>
          </div>
          <Button leftSection={<IconPlayerPlay size={16} />} onClick={handleTrigger} loading={triggering}>
            Run Discovery
          </Button>
        </Group>

        {/* No discovery yet */}
        {!discovery && (
          <Card withBorder p="xl" ta="center">
            <Stack align="center" gap="md">
              <IconChartBar size={48} color="var(--mantine-color-gray-5)" />
              <Title order={3} c="dimmed">No discoveries yet</Title>
              <Text c="dimmed">Run your first discovery to see insights.</Text>
            </Stack>
          </Card>
        )}

        {/* KPI Cards */}
        {discovery && (
          <>
            <Grid>
              <Grid.Col span={{ base: 6, md: 3 }}>
                <Card withBorder p="md" ta="center">
                  <Text size="xl" fw={700} c="blue">{discovery.summary.total_insights}</Text>
                  <Text size="sm" c="dimmed">Insights</Text>
                </Card>
              </Grid.Col>
              <Grid.Col span={{ base: 6, md: 3 }}>
                <Card withBorder p="md" ta="center">
                  <Text size="xl" fw={700} c="violet">{discovery.summary.total_recommendations}</Text>
                  <Text size="sm" c="dimmed">Recommendations</Text>
                </Card>
              </Grid.Col>
              <Grid.Col span={{ base: 6, md: 3 }}>
                <Card withBorder p="md" ta="center">
                  <Text size="xl" fw={700} c="red">
                    {discovery.insights?.filter((i) => i.severity === 'critical').length || 0}
                  </Text>
                  <Text size="sm" c="dimmed">Critical</Text>
                </Card>
              </Grid.Col>
              <Grid.Col span={{ base: 6, md: 3 }}>
                <Card withBorder p="md" ta="center">
                  <Text size="xl" fw={700} c="green">{discovery.summary.queries_executed}</Text>
                  <Text size="sm" c="dimmed">Queries Run</Text>
                </Card>
              </Grid.Col>
            </Grid>

            {/* Insights by Area */}
            {areas.length > 0 && (
              <Tabs defaultValue={areas[0]}>
                <Tabs.List>
                  {areas.map((area) => (
                    <Tabs.Tab key={area} value={area}>
                      {area.charAt(0).toUpperCase() + area.slice(1)} ({insightsByArea[area].length})
                    </Tabs.Tab>
                  ))}
                </Tabs.List>

                {areas.map((area) => (
                  <Tabs.Panel key={area} value={area} pt="md">
                    <Stack gap="md">
                      {insightsByArea[area]
                        .sort((a, b) => b.risk_score - a.risk_score)
                        .map((insight, idx) => (
                          <InsightCard key={idx} insight={insight} />
                        ))}
                    </Stack>
                  </Tabs.Panel>
                ))}
              </Tabs>
            )}

            {/* Recommendations */}
            {discovery.recommendations && discovery.recommendations.length > 0 && (
              <Card withBorder p="lg">
                <Title order={3} mb="md">
                  <IconBulb size={20} style={{ verticalAlign: 'middle', marginRight: 8 }} />
                  Recommendations
                </Title>
                <Stack gap="md">
                  {discovery.recommendations
                    .sort((a, b) => b.priority - a.priority)
                    .map((rec, idx) => (
                      <RecommendationCard key={idx} rec={rec} />
                    ))}
                </Stack>
              </Card>
            )}
          </>
        )}
      </Stack>
    </Shell>
  );
}

function InsightCard({ insight }: { insight: Insight }) {
  return (
    <Card withBorder p="md" radius="md">
      <Group justify="space-between" mb="xs">
        <Group gap="xs">
          <IconAlertTriangle size={16} color={`var(--mantine-color-${severityColor[insight.severity] || 'gray'}-6)`} />
          <Text fw={600}>{insight.name}</Text>
        </Group>
        <Group gap="xs">
          <Badge color={severityColor[insight.severity] || 'gray'} variant="light" size="sm">
            {insight.severity}
          </Badge>
          {insight.affected_count > 0 && (
            <Badge variant="outline" size="sm">{insight.affected_count.toLocaleString()} affected</Badge>
          )}
        </Group>
      </Group>

      <Text size="sm" c="dimmed">{insight.description}</Text>

      {insight.indicators && insight.indicators.length > 0 && (
        <Stack gap={4} mt="sm">
          {insight.indicators.slice(0, 4).map((ind, i) => (
            <Text key={i} size="xs" c="dimmed">- {ind}</Text>
          ))}
        </Stack>
      )}

      {insight.validation && (
        <Badge mt="sm" size="xs" variant="outline"
          color={insight.validation.status === 'confirmed' ? 'green' : insight.validation.status === 'adjusted' ? 'yellow' : 'red'}>
          {insight.validation.status}
        </Badge>
      )}
    </Card>
  );
}

function RecommendationCard({ rec }: { rec: Recommendation }) {
  const priorityColor = rec.priority >= 5 ? 'red' : rec.priority >= 4 ? 'orange' : 'blue';

  return (
    <Card withBorder p="md" radius="md" style={{ borderLeft: `4px solid var(--mantine-color-${priorityColor}-6)` }}>
      <Group justify="space-between" mb="xs">
        <Text fw={600}>{rec.title}</Text>
        <Badge color={priorityColor} variant="light" size="sm">P{rec.priority}</Badge>
      </Group>

      <Text size="sm" c="dimmed" mb="sm">{rec.description}</Text>

      {rec.expected_impact && (
        <Group gap="xs" mb="sm">
          <IconTrendingUp size={14} />
          <Text size="xs" c="dimmed">
            {rec.expected_impact.metric}: {rec.expected_impact.estimated_improvement}
          </Text>
        </Group>
      )}

      {rec.actions && rec.actions.length > 0 && (
        <Stack gap={4}>
          {rec.actions.slice(0, 3).map((action, i) => (
            <Text key={i} size="xs" c="dimmed">- {action}</Text>
          ))}
        </Stack>
      )}
    </Card>
  );
}
