import { useLocation } from 'react-router-dom';
import { JsonConverter } from '../JsonConverter';

export default function OptimizeScript() {
  const location = useLocation<{ sample?: string } | undefined>();
  return <JsonConverter key={location.key} initial={location.state?.sample} />;
}
